package generic

import (
	"reflect"
	"testing"
)

func TestParseTypeMap(t *testing.T) {
	for _, test := range []struct {
		in   string
		want map[string]Target
	}{
		{
			in: "",
		},
		{
			in: "T->V",
		},
		{
			in: "Type->",
		},
		{
			in: " Type  -> OtherType   ",
			want: map[string]Target{
				"Type": Target{
					Ident: "OtherType",
				},
			},
		},
		{
			in: "Type->:OtherType",
		},
		{
			in: "Type->github.com/go:",
		},
		{
			in: "Type->  github.com/go :  go.OtherType ",
			want: map[string]Target{
				"Type": Target{
					Import: "github.com/go",
					Ident:  "go.OtherType",
				},
			},
		},
	} {
		tm, err := ParseTypeMap([]string{test.in})
		if test.want == nil {
			if err == nil {
				t.Fatalf("expect error, got %v", tm)
			}
		} else {
			if !reflect.DeepEqual(tm, test.want) {
				t.Fatalf("expect %v, got %s", test.want, err)
			}
		}
	}
}
