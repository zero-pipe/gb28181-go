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
	ServerGBID  string
	ServerIP    string
	ServerPort  int
	DeviceGBID  string // local platform identity toward upstream
	Username    string // Digest username; defaults to DeviceGBID
	Password    string // Digest password for upstream REGISTER
	Expires     string
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
	to := &sip.FromHeader{Address: recipient}
	req.AppendHeader(from)
	req.AppendHeader(to)
	return req
}

// Register sends REGISTER and completes Digest challenge when Password is set.
func (c *Client) Register(u Upstream) error {
	expires := 3600
	if u.Expires != "" {
		if v, err := strconv.Atoi(u.Expires); err == nil {
			expires = v
		}
	}
	req := c.buildRegister(u, expires)
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
		return fmt.Errorf("cascade REGISTER status %d", res.StatusCode)
	}
	user := u.Username
	if user == "" {
		user = u.DeviceGBID
	}
	if u.Password == "" {
		return fmt.Errorf("cascade REGISTER challenged (%d) but password empty", res.StatusCode)
	}
	authReq := c.buildRegister(u, expires)
	res2, err := c.client.DoDigestAuth(ctx, authReq, res, sipgo.DigestAuth{
		Username: user,
		Password: u.Password,
	})
	if err != nil {
		return err
	}
	if res2.StatusCode != sip.StatusOK {
		return fmt.Errorf("cascade REGISTER auth status %d", res2.StatusCode)
	}
	return nil
}

func (c *Client) Keepalive(u Upstream) error {
	c.sn++
	body := manscdp.BuildPlatformKeepalive(u.DeviceGBID, strconv.Itoa(c.sn))
	recipient := sip.Uri{User: u.ServerGBID, Host: u.ServerIP, Port: u.ServerPort}
	req := sip.NewRequest(sip.MESSAGE, recipient)
	req.SetBody([]byte(body))
	req.AppendHeader(sip.NewHeader("Content-Type", "Application/MANSCDP+xml"))
	from := &sip.FromHeader{Address: sip.Uri{User: u.DeviceGBID, Host: c.cfg.Domain}}
	to := &sip.FromHeader{Address: recipient}
	req.AppendHeader(from)
	req.AppendHeader(to)
	return c.client.WriteRequest(req)
}

func (c *Client) SendCatalogNotify(u Upstream, items []manscdp.CatalogItem) error {
	c.sn++
	body := manscdp.BuildCatalogNotify(u.DeviceGBID, strconv.Itoa(c.sn), items)
	recipient := sip.Uri{User: u.ServerGBID, Host: u.ServerIP, Port: u.ServerPort}
	req := sip.NewRequest(sip.MESSAGE, recipient)
	req.SetBody([]byte(body))
	req.AppendHeader(sip.NewHeader("Content-Type", "Application/MANSCDP+xml"))
	from := &sip.FromHeader{Address: sip.Uri{User: u.DeviceGBID, Host: c.cfg.Domain}}
	to := &sip.FromHeader{Address: recipient}
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
	authReq := c.buildRegister(u, 0)
	res2, err := c.client.DoDigestAuth(ctx, authReq, res, sipgo.DigestAuth{
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
