package client

import (
	"crypto/tls"
	"net/http"
	"time"

	"google.golang.org/grpc"

	"go.ligato.io/vpp-agent/v2/cmd/agentctl/client/tlsconfig"
)

type Opt func(*Client) error

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
		return nil
	}
}

// WithHTTPPort overrides port for HTTP connection.
func WithHTTPPort(p int) Opt {
	return func(c *Client) error {
		c.httpPort = p
		return nil
	}
}

// WithGrpcPort overrides port for GRPC connection.
func WithGrpcPort(p int) Opt {
	return func(c *Client) error {
		c.grpcPort = p
		return nil
	}
}

func WithEtcdEndpoints(endpoints []string) Opt {
	return func(c *Client) error {
		if len(endpoints) != 0 {
			c.kvdbEndpoints = endpoints
		}
		return nil
	}
}

func withTLS(cert, key, ca string, skipVerify bool) (*tls.Config, error) {
	var options []tlsconfig.Option

	if cert != "" && key != "" {
		options = append(options, tlsconfig.CertKey(cert, key))
	}
	if ca != "" {
		options = append(options, tlsconfig.CA(ca))
	}
	if skipVerify {
		options = append(options, tlsconfig.SkipServerVerification())
	}

	return tlsconfig.New(options...)
}

// WithGrpcTLS adds tls.Config for gRPC to Client.
func WithGrpcTLS(cert, key, ca string, skipVerify bool) Opt {
	return func(c *Client) (err error) {
		c.grpcTLS, err = withTLS(cert, key, ca, skipVerify)
		return err
	}
}

// WithHTTPTLS adds tls.Config for HTTP to Client.
func WithHTTPTLS(cert, key, ca string, skipVerify bool) Opt {
	return func(c *Client) (err error) {
		c.httpTLS, err = withTLS(cert, key, ca, skipVerify)
		c.scheme = "https"
		return err
	}
}

// WithKvdbTLS adds tls.Config for KVDB to Client.
func WithKvdbTLS(cert, key, ca string, skipVerify bool) Opt {
	return func(c *Client) (err error) {
		c.kvdbTLS, err = withTLS(cert, key, ca, skipVerify)
		return err
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
		c.HTTPClient().Timeout = timeout
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
