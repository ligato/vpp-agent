package grpcweb

import (
	"bytes"
	"context"
	"io"

	"github.com/ktr0731/grpc-web-go-client/grpcweb/transport"
	"github.com/pkg/errors"
)

type ClientStream interface {
	Send(ctx context.Context, req interface{}) error
	CloseAndReceive(ctx context.Context, res interface{}) error
}

type clientStream struct {
	endpoint    string
	transport   transport.ClientStreamTransport
	callOptions *callOptions
}

func (s *clientStream) Send(ctx context.Context, req interface{}) error {
	r, err := parseRequestBody(s.callOptions.codec, req)
	if err != nil {
		return errors.Wrap(err, "failed to build the request")
	}
	if err := s.transport.Send(ctx, r); err != nil {
		return errors.Wrap(err, "failed to send the request")
	}
	return nil
}

func (s *clientStream) CloseAndReceive(ctx context.Context, res interface{}) error {
	if err := s.transport.CloseSend(); err != nil {
		return errors.Wrap(err, "failed to close the send stream")
	}
	rawBody, err := s.transport.Receive(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to receive the response")
	}
	resBody, err := parseResponseBody(rawBody)
	if err != nil {
		return errors.Wrap(err, "failed to parse the response body")
	}

	if err := s.callOptions.codec.Unmarshal(resBody, res); err != nil {
		return errors.Wrap(err, "failed to unmarshal the response body")
	}
	return nil
}

type ServerStream interface {
	Send(ctx context.Context, req interface{}) error
	Receive(ctx context.Context, res interface{}) error
}

type serverStream struct {
	endpoint    string
	transport   transport.UnaryTransport
	resStream   io.ReadCloser
	callOptions *callOptions
}

func (s *serverStream) Send(ctx context.Context, req interface{}) error {
	codec := s.callOptions.codec

	r, err := parseRequestBody(codec, req)
	if err != nil {
		return errors.Wrap(err, "failed to build the request body")
	}

	contentType := "application/grpc-web+" + codec.Name()
	rawBody, err := s.transport.Send(ctx, s.endpoint, contentType, r)
	if err != nil {
		return errors.Wrap(err, "failed to send the request")
	}
	s.resStream = rawBody
	return nil
}

func (s *serverStream) Receive(ctx context.Context, res interface{}) (err error) {
	if s.resStream == nil {
		return errors.New("Receive must be call after calling Send")
	}
	defer func() {
		if err == io.EOF {
			if rerr := s.transport.Close(); rerr != nil {
				err = rerr
			}
			s.resStream.Close()
		}
	}()

	resBody, err := parseResponseBody(s.resStream)
	if err == io.EOF {
		return io.EOF
	}
	if err != nil {
		return errors.Wrap(err, "failed to parse the response body")
	}

	// check compressed flag.
	// compressed flag is 0 or 1.
	if resBody[0]>>3 != 0 && resBody[0]>>3 != 1 {
		return io.EOF
	}

	if err := s.callOptions.codec.Unmarshal(resBody, res); err != nil {
		return errors.Wrap(err, "failed to unmarshal response body")
	}
	return nil
}

type BidiStream interface {
	Send(ctx context.Context, req interface{}) error
	Receive(ctx context.Context, res interface{}) error
	CloseSend() error
}

type bidiStream struct {
	*clientStream
}

var (
	canonicalGRPCStatusBytes = []byte("Grpc-Status: ")
	gRPCStatusBytes          = []byte("grpc-status: ")
)

func (s *bidiStream) Receive(ctx context.Context, res interface{}) error {
	rawBody, err := s.transport.Receive(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to receive the response")
	}
	resBody, err := parseResponseBody(rawBody)
	if err != nil {
		return errors.Wrap(err, "failed to parse the response body")
	}

	// If trailers appeared, notify it by returning io.EOF.
	if bytes.HasPrefix(resBody, gRPCStatusBytes) || bytes.HasPrefix(resBody, canonicalGRPCStatusBytes) {
		if err := s.transport.Close(); err != nil {
			return errors.Wrap(err, "failed to close the gRPC transport")
		}
		return io.EOF
	}

	if err := s.callOptions.codec.Unmarshal(resBody, res); err != nil {
		return errors.Wrap(err, "failed to unmarshal the response body")
	}
	return nil
}

func (s *bidiStream) CloseSend() error {
	if err := s.transport.CloseSend(); err != nil {
		return errors.Wrap(err, "failed to close the send stream")
	}
	return nil
}
