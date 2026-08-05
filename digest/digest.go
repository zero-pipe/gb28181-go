package digest

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

var authHeaderRe = regexp.MustCompile(`(\w+)=("([^"]*)"|([^,]*))`)

// Challenge is a SIP Digest WWW-Authenticate challenge.
type Challenge struct {
	Realm string
	Nonce string
}

// NewChallenge creates a challenge with a random nonce.
func NewChallenge(realm string) Challenge {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return Challenge{
		Realm: realm,
		Nonce: hex.EncodeToString(b),
	}
}

func (c Challenge) String() string {
	return fmt.Sprintf(`Digest realm="%s", nonce="%s", algorithm=MD5`, c.Realm, c.Nonce)
}

// ParseAuthorization parses a Digest Authorization header into key/value pairs.
func ParseAuthorization(header string) map[string]string {
	result := make(map[string]string)
	for _, match := range authHeaderRe.FindAllStringSubmatch(header, -1) {
		if len(match) < 2 {
			continue
		}
		key := match[1]
		val := match[3]
		if val == "" {
			val = strings.TrimSpace(match[4])
		}
		result[key] = val
	}
	return result
}

// Verify checks a Digest Authorization header against the expected credentials.
// Supports both classic MD5 and qop=auth.
func Verify(authHeader, method, uri, username, password, realm, nonce string) bool {
	auth := ParseAuthorization(authHeader)
	if auth["response"] == "" {
		return false
	}
	if auth["nonce"] != "" {
		nonce = auth["nonce"]
	}
	if auth["realm"] != "" {
		realm = auth["realm"]
	}
	user := auth["username"]
	if user == "" {
		user = username
	}
	digestURI := auth["uri"]
	if digestURI == "" {
		digestURI = uri
	}

	ha1 := md5Hex(fmt.Sprintf("%s:%s:%s", user, realm, password))
	ha2 := md5Hex(fmt.Sprintf("%s:%s", strings.ToUpper(method), digestURI))

	kd := ha1 + ":" + nonce
	if qop := strings.TrimSpace(auth["qop"]); qop != "" {
		if i := strings.Index(qop, ","); i >= 0 {
			qop = strings.TrimSpace(qop[:i])
		}
		if strings.EqualFold(qop, "auth") {
			if nc := auth["nc"]; nc != "" {
				kd += ":" + nc
			}
			if cnonce := auth["cnonce"]; cnonce != "" {
				kd += ":" + cnonce
			}
			kd += ":" + qop
		}
	}
	kd += ":" + ha2
	expected := md5Hex(kd)
	return strings.EqualFold(expected, auth["response"])
}

func md5Hex(s string) string {
	sum := md5.Sum([]byte(s))
	return hex.EncodeToString(sum[:])
}

// WWWAuthenticateHeader builds an HTTP header map with WWW-Authenticate.
func WWWAuthenticateHeader(ch Challenge) http.Header {
	h := http.Header{}
	h.Set("WWW-Authenticate", ch.String())
	return h
}
