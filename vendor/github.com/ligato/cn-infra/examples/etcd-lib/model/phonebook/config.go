package phonebook

import "strings"

//go:generate protoc --proto_path=. --go_out=plugins=grpc:. phonebook.proto

// EtcdPath returns the base path were the phonebook records are stored.
func EtcdPath() string {
	return "/phonebook/"
}

// EtcdContactPath returns the path for a given contact.
func EtcdContactPath(contact *Contact) string {
	return EtcdPath() + strings.Replace(contact.Name, " ", "", -1)
}
