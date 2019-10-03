package transport

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
)

type UnaryTransport interface {
	Send(ctx context.Context, endpoint, contentType string, body io.Reader) (io.ReadCloser, error)
	Close() error
}

type httpTransport struct {
	host   string
	client *http.Client
	opts   *ConnectOptions

	sent bool
}

func (t *httpTransport) Send(ctx context.Context, endpoint, contentType string, body io.Reader) (io.ReadCloser, error) {
	if t.sent {
		return nil, errors.New("Send must be called only one time per one Request")
	}
	defer func() {
		t.sent = true
	}()

	// TODO: HTTPS support.
	scheme := "http"
	u := url.URL{Scheme: scheme, Host: t.host, Path: endpoint}
	url := u.String()
	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to build the API request")
	}

	req.Header.Add("content-type", contentType)
	req.Header.Add("x-grpc-web", "1")

	res, err := t.client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to send the API")
	}

	return res.Body, nil
}

func (t *httpTransport) Close() error {
	t.client.CloseIdleConnections()
	return nil
}

func NewUnary(host string, opts *ConnectOptions) UnaryTransport {
	return &httpTransport{
		host:   host,
		client: http.DefaultClient,
		opts:   opts,
	}
}

type ClientStreamTransport interface {
	Send(ctx context.Context, body io.Reader) error
	Receive(ctx context.Context) (io.ReadCloser, error)

	// CloseSend sends a close signal to the server.
	CloseSend() error

	// Close closes the connection.
	Close() error
}

// webSocketTransport is a stream transport implementation.
//
// Currently, gRPC-Web specification does not support client streaming. (https://github.com/improbable-eng/grpc-web#client-side-streaming)
// webSocketTransport supports improbable-eng/grpc-web's own implementation.
//
// spec: https://github.com/grpc/grpc/blob/master/doc/PROTOCOL-WEB.md
type webSocketTransport struct {
	host     string
	endpoint string

	conn *websocket.Conn

	once    sync.Once
	resOnce sync.Once

	closed bool

	writeMu sync.Mutex
}

func (t *webSocketTransport) Send(ctx context.Context, body io.Reader) error {
	if t.closed {
		return io.EOF
	}

	var err error
	t.once.Do(func() {
		h := http.Header{}
		h.Set("content-type", "application/grpc-web+proto")
		h.Set("x-grpc-web", "1")
		var b bytes.Buffer
		h.Write(&b)

		t.writeMessage(websocket.BinaryMessage, b.Bytes())
	})
	if err != nil {
		return err
	}

	var b bytes.Buffer
	b.Write([]byte{0x00})
	_, err = io.Copy(&b, body)
	if err != nil {
		return errors.Wrap(err, "failed to read request body")
	}

	return t.writeMessage(websocket.BinaryMessage, b.Bytes())
}

func (t *webSocketTransport) Receive(context.Context) (res io.ReadCloser, err error) {
	if t.closed {
		return nil, io.EOF
	}

	defer func() {
		if err == nil {
			return
		}

		if berr, ok := errors.Cause(err).(*net.OpError); ok && !berr.Temporary() {
			err = io.EOF
		}
	}()

	// skip response header
	t.resOnce.Do(func() {
		_, _, err = t.conn.ReadMessage()
		if err != nil {
			err = errors.Wrap(err, "failed to read response header")
			return
		}

		_, _, err = t.conn.ReadMessage()
		if err != nil {
			err = errors.Wrap(err, "failed to read response header")
			return
		}
	})

	var buf bytes.Buffer
	var b []byte

	_, b, err = t.conn.ReadMessage()
	if err != nil {
		if cerr, ok := err.(*websocket.CloseError); ok {
			if cerr.Code == websocket.CloseNormalClosure {
				return nil, io.EOF
			}
		}
		err = errors.Wrap(err, "failed to read response body")
		return
	}
	buf.Write(b)

	var r io.Reader
	_, r, err = t.conn.NextReader()
	if err != nil {
		return
	}

	res = ioutil.NopCloser(io.MultiReader(&buf, r))

	return
}

func (t *webSocketTransport) CloseSend() error {
	// 0x01 means the finish send frame.
	// ref. transports/websocket/websocket.ts
	t.writeMessage(websocket.BinaryMessage, []byte{0x01})
	return nil
}

func (t *webSocketTransport) Close() error {
	// Send the close message.
	err := t.writeMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		return err
	}
	t.closed = true
	// Close the WebSocket connection.
	return t.conn.Close()
}

func (t *webSocketTransport) writeMessage(msg int, b []byte) error {
	t.writeMu.Lock()
	defer t.writeMu.Unlock()
	return t.conn.WriteMessage(msg, b)
}

func NewClientStream(host, endpoint string) (ClientStreamTransport, error) {
	// TODO: WebSocket over TLS support.
	u := url.URL{Scheme: "ws", Host: host, Path: endpoint}
	h := http.Header{}
	h.Set("Sec-WebSocket-Protocol", "grpc-websockets")
	var conn *websocket.Conn
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), h)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to dial to '%s'", u.String())
	}

	return &webSocketTransport{
		host:     host,
		endpoint: endpoint,
		conn:     conn,
	}, nil
}
