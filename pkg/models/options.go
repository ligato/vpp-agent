package models

import (
	"net"
	"strings"
	"text/template"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"

	"go.ligato.io/vpp-agent/v3/proto/ligato/generic"
)

type modelOptions struct {
	nameTemplate string
	nameFunc     NameFunc
}

// ModelOption defines function type which sets model options.
type ModelOption func(*modelOptions)

// NameFunc represents function which can name model instance.
type NameFunc func(x any) (string, error)

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
	return func(x any) (string, error) {
		// handling locally known dynamic messages (they don't have data fields as generated proto messages)
		// (dynamic messages of remotely known models are not supported, remote_model implementation is
		// not using dynamic message for name template resolving so it is ok)
		if dynMessage, ok := x.(*dynamicpb.Message); ok {
			var err error
			x, err = resolveDynamicProtoModelName(dynMessage)
			if err != nil {
				return "", err
			}
		}

		// execute name template on generated proto message
		var s strings.Builder
		if err := tmpl.Execute(&s, x); err != nil {
			return "", err
		}
		return s.String(), nil
	}
}

func defaultOptions(x any) modelOptions {
	var opts modelOptions
	if _, ok := x.(named); ok {
		opts.nameFunc = func(x any) (s string, err error) {
			// handling dynamic messages (they don't implement named interface)
			if dynMessage, ok := x.(*dynamicpb.Message); ok {
				x, err = DynamicLocallyKnownMessageToGeneratedMessage(dynMessage)
				if err != nil {
					return "", err
				}
			}
			// handling other proto message
			return x.(named).GetName(), nil
		}
		opts.nameTemplate = namedTemplate
	}
	return opts
}

func optsFromProtoDesc(desc protoreflect.MessageDescriptor) []ModelOption {
	var opts []ModelOption
	descOpts := desc.Options()
	if proto.HasExtension(descOpts, generic.E_ModelNameTemplate) {
		t := proto.GetExtension(descOpts, generic.E_ModelNameTemplate).(string)
		opts = append(opts, WithNameTemplate(t))
	}
	return opts
}

var funcMap = template.FuncMap{
	"field": func(msg proto.Message, fieldNum protoreflect.FieldNumber) string {
		desc := msg.ProtoReflect().Descriptor().Fields().ByNumber(fieldNum)
		if desc == nil {
			return "<invalid>"
		}
		return msg.ProtoReflect().Get(desc).String()
	},
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
