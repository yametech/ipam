package client

import (
	"encoding/json"
	"net"
	"net/http"

	typesVer "github.com/containernetworking/cni/pkg/types/100"
	"github.com/go-resty/resty/v2"
	store "github.com/yametech/global-ipam/pkg/cni"
	"github.com/yametech/global-ipam/pkg/config"
	"github.com/yametech/global-ipam/pkg/dns"
)

var _ store.Cni = &Client{}

type CniClient struct {
	*resty.Client
}

type Client struct {
	cli *CniClient
}

func New() store.Cni {
	return &Client{
		cli: NewCniClient(store.UNIX_SOCK_PATH),
	}
}

func (c *Client) Allocate(namespace, pod string, cfg *config.Net) (*typesVer.Result, error) {
	cfgBytes, err := json.Marshal(cfg)
	if err != nil {
		return nil, err
	}
	r := &store.AllocateResponse{}
	_, err = c.cli.R().
		SetHeader("Content-Type", "application/json").
		SetFormData(map[string]string{
			"namespace": namespace,
			"pod":       pod,
			"cfg":       string(cfgBytes),
		}).
		SetResult(r).
		Post("/allocate")

	if err != nil {
		return nil, err
	}

	result := r.Result
	if cfg.ResolvConf != "" {
		dns, err := dns.ParseResolvConf(cfg.ResolvConf)
		if err != nil {
			return nil, err
		}
		result.DNS = *dns
	}

	return result, nil
}

func (c *Client) Release(namespace, pod string) error {
	_, err := c.cli.R().
		SetHeader("Content-Type", "application/json").
		SetFormData(map[string]string{
			"namespace": namespace,
			"pod":       pod,
		}).
		Post("/release")
	if err != nil {
		return err
	}
	return nil
}

func NewCniClient(socketAddress string) *CniClient {
	transport := http.Transport{
		Dial: func(_, _ string) (net.Conn, error) {
			return net.Dial("unix", store.UNIX_SOCK_PATH)
		},
	}
	// Create a Resty Client
	client := resty.NewWithClient(&http.Client{Transport: &transport}).
		SetScheme("http").
		SetHostURL("http://dummy")

	return &CniClient{Client: client}
}
