package digest

import (
	"crypto/md5"
	"encoding/hex"
	"testing"
)

func md5HexRef(s string) string {
	sum := md5.Sum([]byte(s))
	return hex.EncodeToString(sum[:])
}

func TestVerifyWithoutQop(t *testing.T) {
	user := "34020000001320000001"
	realm := "3402000000"
	pass := "12345678"
	method := "REGISTER"
	uri := "sip:34020000001320000001@3402000000"
	nonce := "deadbeef"

	ha1 := md5HexRef(user + ":" + realm + ":" + pass)
	ha2 := md5HexRef(method + ":" + uri)
	response := md5HexRef(ha1 + ":" + nonce + ":" + ha2)

	hdr := `Digest username="` + user + `", realm="` + realm + `", nonce="` + nonce +
		`", uri="` + uri + `", response="` + response + `"`

	if !Verify(hdr, method, "sip:wrong@host", user, pass, realm, nonce) {
		t.Fatal("expected digest without qop to pass using uri from Authorization header")
	}
}

func TestVerifyWithQopAuth(t *testing.T) {
	user := "34020000001320000001"
	realm := "3402000000"
	pass := "12345678"
	method := "REGISTER"
	uri := "sip:34020000001320000001@3402000000"
	nonce := "abc123"
	nc := "00000001"
	cnonce := "0a4f113b"
	qop := "auth"

	ha1 := md5HexRef(user + ":" + realm + ":" + pass)
	ha2 := md5HexRef(method + ":" + uri)
	response := md5HexRef(ha1 + ":" + nonce + ":" + nc + ":" + cnonce + ":" + qop + ":" + ha2)

	hdr := `Digest username="` + user + `", realm="` + realm + `", nonce="` + nonce +
		`", uri="` + uri + `", response="` + response + `", qop=` + qop +
		`, nc=` + nc + `, cnonce="` + cnonce + `"`

	if !Verify(hdr, method, "sip:other@192.168.1.5:8116", user, pass, realm, nonce) {
		t.Fatal("expected digest with qop=auth to pass")
	}
}
