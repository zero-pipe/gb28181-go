package cascade

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/emiago/sipgo"
	"github.com/emiago/sipgo/sip"
	"github.com/zero-pipe/gb28181-go/manscdp"
)

// Upstream describes a superior GB28181 platform.
type Upstream struct {
	ServerGBID string
	ServerIP   string
	ServerPort int
	DeviceGBID string // local platform identity toward upstream
	Username   string // Digest username; defaults to DeviceGBID
	Password   string // Digest password for upstream REGISTER
	Expires    string
}

// Config is local SIP identity used in From headers.
type Config struct {
	ID     string
	Domain string
}

// Client registers / keepalives toward an upstream platform.
type Client struct {
	cfg    Config
	ua     *sipgo.UserAgent
	client *sipgo.Client
	sn     int
}

// NewClient creates an upstream cascade client.
func NewClient(cfg Config) *Client {
	ua, _ := sipgo.NewUA(sipgo.WithUserAgent(cfg.ID))
	client, _ := sipgo.NewClient(ua)
	return &Client{cfg: cfg, ua: ua, client: client}
}

func (c *Client) ApplyConfig(cfg Config) { c.cfg = cfg }

func (c *Client) buildRegister(u Upstream, expires int) *sip.Request {
	recipient := sip.Uri{User: u.ServerGBID, Host: u.ServerIP, Port: u.ServerPort}
	req := sip.NewRequest(sip.REGISTER, recipient)
	req.AppendHeader(sip.NewHeader("Expires", strconv.Itoa(expires)))
	from := &sip.FromHeader{Address: sip.Uri{User: u.DeviceGBID, Host: c.cfg.Domain}}
	to := &sip.ToHeader{Address: recipient}
	req.AppendHeader(from)
	req.AppendHeader(to)
	return req
}

// Register sends REGISTER and completes Digest challenge when Password is set.
// A 302 Moved Temporarily is followed to the Contact target (2022 §9.1.2.3),
// up to maxRedirects hops.
func (c *Client) Register(u Upstream) error {
	expires := 3600
	if u.Expires != "" {
		if v, err := strconv.Atoi(u.Expires); err == nil {
			expires = v
		}
	}
	const maxRedirects = 3
	for hop := 0; hop <= maxRedirects; hop++ {
		res, err := c.registerOnce(&u, expires)
		if err != nil {
			return err
		}
		switch res.StatusCode {
		case sip.StatusOK:
			return nil
		case sip.StatusMovedTemporarily:
			target, ok := redirectContact(res)
			if !ok {
				return fmt.Errorf("cascade REGISTER 302 without usable Contact")
			}
			u.ServerIP, u.ServerPort, u.ServerGBID = target.host, target.port, target.user
			continue
		default:
			return fmt.Errorf("cascade REGISTER status %d", res.StatusCode)
		}
	}
	return fmt.Errorf("cascade REGISTER exceeded %d redirects", maxRedirects)
}

// registerOnce performs one REGISTER round-trip, handling the Digest challenge.
func (c *Client) registerOnce(u *Upstream, expires int) (*sip.Response, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	req := c.buildRegister(*u, expires)
	res, err := c.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != sip.StatusUnauthorized && res.StatusCode != sip.StatusProxyAuthRequired {
		return res, nil
	}
	user := u.Username
	if user == "" {
		user = u.DeviceGBID
	}
	if u.Password == "" {
		return nil, fmt.Errorf("cascade REGISTER challenged (%d) but password empty", res.StatusCode)
	}
	// DoDigestAuth re-sends the same request object; a fresh request would have
	// no CSeq and panics inside sipgo.
	return c.client.DoDigestAuth(ctx, req, res, sipgo.DigestAuth{
		Username: user,
		Password: u.Password,
	})
}

type redirectTarget struct {
	user string
	host string
	port int
}

// redirectContact extracts the REGISTER redirect target from a 302 Contact.
func redirectContact(res *sip.Response) (redirectTarget, bool) {
	contact := res.Contact()
	if contact == nil {
		return redirectTarget{}, false
	}
	addr := contact.Address
	if addr.Host == "" {
		return redirectTarget{}, false
	}
	port := addr.Port
	if port <= 0 {
		port = 5060
	}
	return redirectTarget{user: addr.User, host: addr.Host, port: port}, true
}

func (c *Client) Keepalive(u Upstream) error {
	c.sn++
	body := manscdp.BuildPlatformKeepalive(u.DeviceGBID, strconv.Itoa(c.sn))
	recipient := sip.Uri{User: u.ServerGBID, Host: u.ServerIP, Port: u.ServerPort}
	req := sip.NewRequest(sip.MESSAGE, recipient)
	req.SetBody(manscdp.EncodeBody(body))
	req.AppendHeader(sip.NewHeader("Content-Type", "Application/MANSCDP+xml"))
	from := &sip.FromHeader{Address: sip.Uri{User: u.DeviceGBID, Host: c.cfg.Domain}}
	to := &sip.ToHeader{Address: recipient}
	req.AppendHeader(from)
	req.AppendHeader(to)
	return c.client.WriteRequest(req)
}

func (c *Client) SendCatalogNotify(u Upstream, items []manscdp.CatalogItem) error {
	c.sn++
	body := manscdp.BuildCatalogNotify(u.DeviceGBID, strconv.Itoa(c.sn), items)
	recipient := sip.Uri{User: u.ServerGBID, Host: u.ServerIP, Port: u.ServerPort}
	req := sip.NewRequest(sip.MESSAGE, recipient)
	req.SetBody(manscdp.EncodeBody(body))
	req.AppendHeader(sip.NewHeader("Content-Type", "Application/MANSCDP+xml"))
	from := &sip.FromHeader{Address: sip.Uri{User: u.DeviceGBID, Host: c.cfg.Domain}}
	to := &sip.ToHeader{Address: recipient}
	req.AppendHeader(from)
	req.AppendHeader(to)
	return c.client.WriteRequest(req)
}

func (c *Client) Unregister(u Upstream) error {
	req := c.buildRegister(u, 0)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	res, err := c.client.Do(ctx, req)
	if err != nil {
		return err
	}
	if res.StatusCode == sip.StatusOK {
		return nil
	}
	if res.StatusCode != sip.StatusUnauthorized && res.StatusCode != sip.StatusProxyAuthRequired {
		return fmt.Errorf("cascade UNREGISTER status %d", res.StatusCode)
	}
	user := u.Username
	if user == "" {
		user = u.DeviceGBID
	}
	if u.Password == "" {
		return fmt.Errorf("cascade UNREGISTER challenged (%d) but password empty", res.StatusCode)
	}
	res2, err := c.client.DoDigestAuth(ctx, req, res, sipgo.DigestAuth{
		Username: user,
		Password: u.Password,
	})
	if err != nil {
		return err
	}
	if res2.StatusCode != sip.StatusOK {
		return fmt.Errorf("cascade UNREGISTER auth status %d", res2.StatusCode)
	}
	return nil
}
