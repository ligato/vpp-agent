package main

import "fmt"

// Package represents collection of objects parsed from VPP binary API JSON data
type Package struct {
	Name     string
	Version  string
	CRC      string
	Services []Service
	Enums    []Enum
	Aliases  []Alias
	Types    []Type
	Unions   []Union
	Messages []Message
	RefMap   map[string]string
}

// Service represents VPP binary API service
type Service struct {
	Name        string
	RequestType string
	ReplyType   string
	Stream      bool
	Events      []string
}

// Enum represents VPP binary API enum
type Enum struct {
	Name    string
	Type    string
	Entries []EnumEntry
}

// EnumEntry represents VPP binary API enum entry
type EnumEntry struct {
	Name  string
	Value interface{}
}

// Alias represents VPP binary API alias
type Alias struct {
	Name   string
	Type   string
	Length int
}

// Type represents VPP binary API type
type Type struct {
	Name   string
	CRC    string
	Fields []Field
}

// Field represents VPP binary API object field
type Field struct {
	Name     string
	Type     string
	Length   int
	SizeFrom string
	Meta     FieldMeta
}

// FieldMeta represents VPP binary API meta info for field
type FieldMeta struct {
	Limit int
}

// Union represents VPP binary API union
type Union struct {
	Name   string
	CRC    string
	Fields []Field
}

// Message represents VPP binary API message
type Message struct {
	Name   string
	CRC    string
	Fields []Field
}

// MessageType represents the type of a VPP message
type MessageType int

const (
	requestMessage MessageType = iota // VPP request message
	replyMessage                      // VPP reply message
	eventMessage                      // VPP event message
	otherMessage                      // other VPP message
)

// printPackage prints all loaded objects for package
func printPackage(pkg *Package) {
	logf("package: %s %s (%s)", pkg.Name, pkg.Version, pkg.CRC)
	if len(pkg.Enums) > 0 {
		logf(" %d enums:", len(pkg.Enums))
		for _, enum := range pkg.Enums {
			logf("  - %s: %+v", enum.Name, enum)
		}
	}
	if len(pkg.Unions) > 0 {
		logf(" %d unions:", len(pkg.Unions))
		for _, union := range pkg.Unions {
			logf("  - %s: %+v", union.Name, union)
		}
	}
	if len(pkg.Types) > 0 {
		logf(" %d types:", len(pkg.Types))
		for _, typ := range pkg.Types {
			logf("  - %s (%d fields): %+v", typ.Name, len(typ.Fields), typ)
		}
	}
	if len(pkg.Messages) > 0 {
		logf(" %d messages:", len(pkg.Messages))
		for _, msg := range pkg.Messages {
			logf("  - %s (%d fields) %s", msg.Name, len(msg.Fields), msg.CRC)
		}
	}
	if len(pkg.Services) > 0 {
		logf(" %d services:", len(pkg.Services))
		for _, svc := range pkg.Services {
			var info string
			if svc.Stream {
				info = "(STREAM)"
			} else if len(svc.Events) > 0 {
				info = fmt.Sprintf("(EVENTS: %v)", svc.Events)
			}
			logf("  - %s: %q -> %q %s", svc.Name, svc.RequestType, svc.ReplyType, info)
		}
	}
}
