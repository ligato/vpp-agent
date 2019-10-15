// Package etcdexample explains how to generate Golang structures from
// protobuf-formatted data.
package etcdexample

//go:generate protoc --proto_path=. --go_out=. example.proto
