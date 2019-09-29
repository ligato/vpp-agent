package client

import (
	"net/http"
	"os"
	"time"

	"github.com/docker/go-connections/sockets"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

type Opt func(*Client) error

// FromEnv configures the client with values from environment variables.
//
// Supported environment variables:
// - AGENT_HOST to set the url to the agent server.
// - LIGATO_API_VERSION to set the url to the agent server.
func FromEnv(c *Client) error {
	if host := os.Getenv("AGENT_HOST"); host != "" {
		if err := WithHost(host)(c); err != nil {
			return err
		}
	}
	if version := os.Getenv("LIGATO_API_VERSION"); version != "" {
		if err := WithVersion(version)(c); err != nil {
			return err
		}
	}
	return nil
}

// WithHost overrides the client host with the specified one.
func WithHost(host string) Opt {
	return func(c *Client) error {
		hostURL, err := ParseHostURL(host)
		if err != nil {
			return err
		}
		c.host = host
		c.proto = hostURL.Scheme
		c.addr = hostURL.Host
		c.basePath = hostURL.Path
		if transport, ok := c.httpClient.Transport.(*http.Transport); ok {
			return sockets.ConfigureTransport(transport, c.proto, c.addr)
		}
		return errors.Errorf("cannot apply host to transport: %T", c.httpClient.Transport)
	}
}

func WithGRPCAddr(addr string) Opt {
	return func(c *Client) error {
		if addr != "" {
			c.grpcAddr = addr
		}
		return nil
	}
}

func WithEtcdEndpoints(endpoints []string) Opt {
	return func(c *Client) error {
		if len(endpoints) != 0 {
			c.endpoints = endpoints
		}
		return nil
	}
}

func WithServiceLabel(label string) Opt {
	return func(c *Client) error {
		if label != "" {
			c.serviceLabel = label
		}
		return nil
	}
}

// WithHTTPClient overrides the http client with the specified one
func WithHTTPClient(client *http.Client) Opt {
	return func(c *Client) error {
		if client != nil {
			c.httpClient = client
		}
		return nil
	}
}

// WithGRPCClient overrides the grpc client with the specified one
func WithGRPCClient(client *grpc.ClientConn) Opt {
	return func(c *Client) error {
		if client != nil {
			c.grpcClient = client
		}
		return nil
	}
}

// WithTimeout configures the time limit for requests made by the HTTP client
func WithTimeout(timeout time.Duration) Opt {
	return func(c *Client) error {
		c.httpClient.Timeout = timeout
		return nil
	}
}

// WithHTTPHeaders overrides the client default http headers
func WithHTTPHeaders(headers map[string]string) Opt {
	return func(c *Client) error {
		c.customHTTPHeaders = headers
		return nil
	}
}

// WithVersion overrides the client version with the specified one. If an empty
// version is specified, the value will be ignored to allow version negotiation.
func WithVersion(version string) Opt {
	return func(c *Client) error {
		if version != "" {
			c.version = version
			c.manualOverride = true
		}
		return nil
	}
}

// WithAPIVersionNegotiation enables automatic API version negotiation for the client.
// With this option enabled, the client automatically negotiates the API version
// to use when making requests. API version negotiation is performed on the first
// request; subsequent requests will not re-negotiate.
func WithAPIVersionNegotiation() Opt {
	return func(c *Client) error {
		c.negotiateVersion = true
		return nil
	}
}
