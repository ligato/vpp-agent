package models

import (
	"net"
	"reflect"
	"strings"
	"text/template"

	"google.golang.org/protobuf/types/dynamicpb"
)

type modelOptions struct {
	nameTemplate string
	nameFunc     NameFunc
}

// ModelOption defines function type which sets model options.
type ModelOption func(*modelOptions)

// NameFunc represents function which can name model instance.
// To properly handle also dynamic Messages (dynamicpb.Message)
// as model instances, the go type of corresponding generated
// proto message must be given.
type NameFunc func(obj interface{}, messageGoType reflect.Type) (string, error)

// WithNameTemplate returns option for models which sets function
// for generating name of instances using custom template.
func WithNameTemplate(t string) ModelOption {
	return func(opts *modelOptions) {
		opts.nameFunc = NameTemplate(t)
		opts.nameTemplate = t
	}
}

const namedTemplate = `{{.Name}}`

type named interface {
	GetName() string
}

func NameTemplate(t string) NameFunc {
	tmpl := template.Must(
		template.New("name").Funcs(funcMap).Option("missingkey=error").Parse(t),
	)
	return func(obj interface{}, messageGoType reflect.Type) (string, error) {
		// handling dynamic messages (they don't have data fields as generated proto messages)
		if dynMessage, ok := obj.(*dynamicpb.Message); ok {
			var err error
			obj, err = dynamicMessageToGeneratedMessage(dynMessage, messageGoType)
			if err != nil {
				return "", err
			}
		}

		// execute name template on generated proto message
		var s strings.Builder
		if err := tmpl.Execute(&s, obj); err != nil {
			return "", err
		}
		return s.String(), nil
	}
}

var funcMap = template.FuncMap{
	"ip": func(s string) string {
		ip := net.ParseIP(s)
		if ip == nil {
			return "<invalid>"
		}
		return ip.String()
	},
	"protoip": func(s string) string {
		ip := net.ParseIP(s)
		if ip == nil {
			return "<invalid>"
		}

		if ip.To4() == nil {
			return "IPv6"
		}
		return "IPv4"
	},
	"ipnet": func(s string) map[string]interface{} {
		if strings.HasPrefix(s, "alloc:") {
			// reference to IP address allocated via netalloc
			return nil
		}
		_, ipNet, err := net.ParseCIDR(s)
		if err != nil {
			return map[string]interface{}{
				"IP":       "<invalid>",
				"MaskSize": 0,
				"AllocRef": "",
			}
		}
		maskSize, _ := ipNet.Mask.Size()
		return map[string]interface{}{
			"IP":       ipNet.IP.String(),
			"MaskSize": maskSize,
			"AllocRef": "",
		}
	},
}
