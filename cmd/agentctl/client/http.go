package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"go.ligato.io/vpp-agent/v3/cmd/agentctl/api/types"
)

// serverResponse is a wrapper for http API responses.
type serverResponse struct {
	body       io.ReadCloser
	header     http.Header
	statusCode int
	reqURL     *url.URL
}

func (c *Client) get(ctx context.Context, path string, query url.Values, headers map[string][]string) (serverResponse, error) {
	return c.sendRequest(ctx, "GET", path, query, nil, headers)
}

func (c *Client) post(ctx context.Context, path string, query url.Values, obj interface{}, headers map[string][]string) (serverResponse, error) {
	body, headers, err := encodeBody(obj, headers)
	if err != nil {
		return serverResponse{}, err
	}
	return c.sendRequest(ctx, "POST", path, query, body, headers)
}

func (c *Client) put(ctx context.Context, path string, query url.Values, obj interface{}, headers map[string][]string) (serverResponse, error) {
	body, headers, err := encodeBody(obj, headers)
	if err != nil {
		return serverResponse{}, err
	}
	return c.sendRequest(ctx, "PUT", path, query, body, headers)
}

type headers map[string][]string

func encodeBody(obj interface{}, headers headers) (io.Reader, headers, error) {
	if obj == nil {
		return nil, headers, nil
	}

	body, err := encodeData(obj)
	if err != nil {
		return nil, headers, err
	}
	if headers == nil {
		headers = make(map[string][]string)
	}
	headers["Content-Type"] = []string{"application/json"}
	return body, headers, nil
}

func (c *Client) buildRequest(method, path string, body io.Reader, headers headers) (*http.Request, error) {
	expectedPayload := method == "POST" || method == "PUT"
	if expectedPayload && body == nil {
		body = bytes.NewReader([]byte{})
	}

	req, err := http.NewRequest(method, path, body)
	if err != nil {
		return nil, err
	}
	req = c.addHeaders(req, headers)

	if c.proto == "unix" || c.proto == "npipe" {
		// For local communications, it doesn't matter what the host is.
		// We just need a valid and meaningful host name.
		req.Host = "ligato-agent"
	}

	req.URL.Host = c.httpAddr
	req.URL.Scheme = c.scheme

	if expectedPayload && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "text/plain")
	}
	return req, nil
}

func (c *Client) sendRequest(ctx context.Context, method, path string, query url.Values, body io.Reader, headers headers) (serverResponse, error) {
	req, err := c.buildRequest(method, c.getAPIPath(ctx, path, query), body, headers)
	if err != nil {
		return serverResponse{}, err
	}

	resp, err := c.doRequest(ctx, req)
	if err != nil {
		return resp, err
	}
	err = c.checkResponseErr(resp)
	return resp, err
}

func (c *Client) doRequest(ctx context.Context, req *http.Request) (serverResponse, error) {
	serverResp := serverResponse{statusCode: -1, reqURL: req.URL}

	var (
		err  error
		resp *http.Response
	)
	req = req.WithContext(ctx)

	logrus.Debugf("=> sending http request: %s %s (%d bytes)", req.Method, req.URL, req.ContentLength)
	defer func() {
		if err != nil {
			logrus.Debugf("<= http response ERROR: %v", err)
		} else {
			logrus.Debugf("<= http response (statusCode %v)", serverResp.statusCode)
		}
	}()

	resp, err = c.HTTPClient().Do(req)
	if err != nil {
		if c.scheme != "https" && strings.Contains(err.Error(), "malformed HTTP response") {
			return serverResp, fmt.Errorf("%v.\n* Are you trying to connect to a TLS-enabled daemon without TLS?", err)
		}

		if c.scheme == "https" && strings.Contains(err.Error(), "bad certificate") {
			return serverResp, errors.Wrap(err, "The server probably has client authentication (--tlsverify) enabled. Please check your TLS client certification settings")
		}

		// Don't decorate context sentinel errors; users may be comparing to
		// them directly.
		switch err {
		case context.Canceled, context.DeadlineExceeded:
			return serverResp, err
		}

		if nErr, ok := err.(*url.Error); ok {
			if nErr, ok := nErr.Err.(*net.OpError); ok {
				if os.IsPermission(nErr.Err) {
					return serverResp, errors.Wrapf(err, "Got permission denied while trying to connect to the agent socket at %v", c.host)
				}
			}
		}

		if err, ok := err.(net.Error); ok {
			if err.Timeout() {
				return serverResp, ErrorConnectionFailed(c.host)
			}
			if !err.Temporary() {
				if strings.Contains(err.Error(), "connection refused") || strings.Contains(err.Error(), "dial unix") {
					return serverResp, ErrorConnectionFailed(c.host)
				}
			}
		}

		return serverResp, errors.Wrap(err, "error during connect")
	}

	if resp != nil {
		serverResp.statusCode = resp.StatusCode
		serverResp.body = resp.Body
		serverResp.header = resp.Header
	}
	return serverResp, nil
}

func (c *Client) checkResponseErr(serverResp serverResponse) error {
	if serverResp.statusCode >= 200 && serverResp.statusCode < 400 {
		return nil
	}

	var body []byte
	var err error
	if serverResp.body != nil {
		bodyMax := 1 * 1024 * 1024 // 1 MiB
		bodyR := &io.LimitedReader{
			R: serverResp.body,
			N: int64(bodyMax),
		}
		body, err = ioutil.ReadAll(bodyR)
		if err != nil {
			return err
		}
		if bodyR.N == 0 {
			return fmt.Errorf("request returned %s with a message (> %d bytes) for API route and version %s, check if the server supports the requested API version",
				http.StatusText(serverResp.statusCode), bodyMax, serverResp.reqURL)
		}
	}
	if len(body) == 0 {
		return fmt.Errorf("request returned %s for API route and version %s, check if the server supports the requested API version",
			http.StatusText(serverResp.statusCode), serverResp.reqURL)
	}

	var ct string
	if serverResp.header != nil {
		ct = serverResp.header.Get("Content-Type")
	}

	var errorMsg string
	if ct == "application/json" {
		var errorResponse types.ErrorResponse
		if err := json.Unmarshal(body, &errorResponse); err != nil {
			return errors.Wrap(err, "Error reading JSON")
		}
		errorMsg = errorResponse.Message
	} else {
		errorMsg = string(body)
	}

	errorMsg = fmt.Sprintf("[%d] %s", serverResp.statusCode, strings.TrimSpace(errorMsg))

	return errors.Wrap(errors.New(errorMsg), "Error response from daemon")
}

func (c *Client) addHeaders(req *http.Request, headers headers) *http.Request {
	// Add CLI Config's HTTP Headers BEFORE we set the client headers
	// then the user can't change OUR headers
	for k, v := range c.customHTTPHeaders {
		req.Header.Set(k, v)
	}
	if headers != nil {
		for k, v := range headers {
			req.Header[k] = v
		}
	}
	return req
}

func encodeData(data interface{}) (*bytes.Buffer, error) {
	params := bytes.NewBuffer(nil)
	if data != nil {
		if err := json.NewEncoder(params).Encode(data); err != nil {
			return nil, err
		}
	}
	return params, nil
}

func ensureReaderClosed(response serverResponse) {
	if response.body != nil {
		// Drain up to 512 bytes and close the body to let the Transport reuse the connection
		_, _ = io.CopyN(ioutil.Discard, response.body, 512)
		_ = response.body.Close()
	}
}
