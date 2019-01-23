package kvscheduler

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/ligato/vpp-agent/plugins/kvscheduler/api"
	"github.com/ligato/vpp-agent/plugins/kvscheduler/internal/graph"
	"github.com/unrolled/render"
)

func (s *Scheduler) dotGraphHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		graphRead := s.graph.Read()
		defer graphRead.Release()

		output, err := renderDotOutput(graphRead)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Printf("DOT:\n%s\n", output)

		img, err := dotToImage("", "svg", output)
		if err != nil {
			http.Error(w, fmt.Sprintf("%v\n%v", err, img), http.StatusInternalServerError)
			return
		}

		log.Println("serving file:", img)
		http.ServeFile(w, req, img)
	}
}

func renderDotOutput(g graph.ReadAccess) ([]byte, error) {
	cluster := NewDotCluster("focus")
	cluster.Attrs = dotAttrs{
		"bgcolor":   "white",
		"label":     "",
		"labelloc":  "t",
		"labeljust": "c",
		"fontsize":  "18",
		"tooltip":   "",
	}
	/*if focusPkg != nil {
		cluster.Attrs["bgcolor"] = "#e6ecfa"
		cluster.Attrs["label"] = focusPkg.Name
	}*/

	var (
		nodes []*dotNode
		edges []*dotEdge
	)

	nodeMap := make(map[string]*dotNode)
	edgeMap := make(map[string]*dotEdge)

	var processGraphNode = func(graphNode graph.Node) *dotNode {
		key := graphNode.GetKey()

		if n, ok := nodeMap[key]; ok {
			return n
		}
		attrs := make(dotAttrs)

		fmt.Printf("- key: %q\n", key)

		if label := graphNode.GetLabel(); label != "" {
			attrs["label"] = label
		}

		c := cluster

		descriptorFlag := graphNode.GetFlag(DescriptorFlagName)
		if descriptorFlag != nil {
			attrs["fillcolor"] = "PaleGreen"

			descriptor := descriptorFlag.GetValue()
			if _, ok := c.Clusters[descriptor]; !ok {
				c.Clusters[descriptor] = &dotCluster{
					ID:       key,
					Clusters: make(map[string]*dotCluster),
					Attrs: dotAttrs{
						"penwidth":  "0.8",
						"fontsize":  "16",
						"label":     fmt.Sprintf("[ %s ]", descriptor),
						"style":     "filled",
						"fillcolor": "#e6ecfa",
						//"fontname":  "bold",
						//"rank":      "sink",
					},
				}
			}
			c = c.Clusters[descriptor]
		}
		origin := graphNode.GetFlag(OriginFlagName)
		if origin != nil {
			if o := origin.GetValue(); o == api.FromSB.String() {
				attrs["fillcolor"] = "LightCyan"
			} else if o == api.FromNB.String() {
				//attrs["penwidth"] = "1.5"
			}
		}
		pending := graphNode.GetFlag(PendingFlagName)
		if pending != nil {
			attrs["style"] = "dashed,filled"
			attrs["fillcolor"] = "Pink"
		}
		//attrs["margin"] = "0.04,0.01"
		//attrs["pad"] = "0.04"

		n := &dotNode{
			ID:    key,
			Attrs: attrs,
		}
		c.Nodes = append(c.Nodes, n)
		nodeMap[key] = n
		return n
	}

	var addEdge = func(e *dotEdge) {
		edgeKey := fmt.Sprintf("%s->%s", e.From.ID, e.To.ID)
		if _, ok := edgeMap[edgeKey]; !ok {
			edges = append(edges, e)
			edgeMap[edgeKey] = e
		}
	}

	for _, key := range g.GetKeys() {
		graphNode := g.GetNode(key)

		n := processGraphNode(graphNode)

		for _, target := range graphNode.GetTargets(DerivesRelation) {
			for _, derivesNode := range target.Nodes {
				d := processGraphNode(derivesNode)
				d.Attrs["fillcolor"] = "LightYellow"
				d.Attrs["style"] = "rounded,filled"
				attrs := make(dotAttrs)
				attrs["color"] = "DarkKhaki"
				attrs["arrowhead"] = "invempty"
				e := &dotEdge{
					From:  n,
					To:    d,
					Attrs: attrs,
				}
				addEdge(e)
			}
		}
		for _, target := range graphNode.GetTargets(DependencyRelation) {
			for _, depNode := range target.Nodes {
				d := processGraphNode(depNode)
				attrs := make(dotAttrs)
				attrs["tooltip"] = target.Label
				e := &dotEdge{
					From:  n,
					To:    d,
					Attrs: attrs,
				}
				addEdge(e)
			}
		}
	}

	hostname, _ := os.Hostname()
	title := fmt.Sprintf("KVScheduler Graph: %d keys - generated %s on %s (PID: %d)",
		len(g.GetKeys()), time.Now().Format(time.RFC1123), hostname, os.Getpid())

	dot := &dotGraph{
		Title:   title,
		Minlen:  minlen,
		Cluster: cluster,
		Nodes:   nodes,
		Edges:   edges,
		Options: map[string]string{
			"minlen":  fmt.Sprint(minlen),
			"nodesep": fmt.Sprint(nodesep),
		},
	}

	var buf bytes.Buffer
	if err := WriteDot(&buf, dot); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

var (
	minlen  uint    = 1
	nodesep float64 = 1
)

// location of dot executable for converting from .dot to .svg
// it's usually at: /usr/bin/dot
var dotExe string

// dotToImage generates a SVG using the 'dot' utility, returning the filepath
func dotToImage(outfname string, format string, dot []byte) (string, error) {
	if dotExe == "" {
		dot, err := exec.LookPath("dot")
		if err != nil {
			return "", fmt.Errorf("unable to find program 'dot', please install it or check your PATH")
		}
		dotExe = dot
	}

	var img string
	if outfname == "" {
		img = filepath.Join(os.TempDir(), fmt.Sprintf("kvscheduler-graph.%s", format))
	} else {
		img = fmt.Sprintf("%s.%s", outfname, format)
	}

	cmd := exec.Command(dotExe, fmt.Sprintf("-T%s", format), "-o", img)
	cmd.Stdin = bytes.NewReader(dot)
	if out, err := cmd.CombinedOutput(); err != nil {
		return string(out), err
	}

	return img, nil
}

const tmplGraph = `digraph kvscheduler {
	ranksep=.5
	//nodesep=.1
    label="{{.Title}}";
	labelloc="t";
    labeljust="l";
    fontsize="12";
	fontname="Ubuntu"; 
    rankdir="LR";
    bgcolor="lightgray";
    style="solid";
    penwidth="1";
    pad="0.05";
    nodesep="{{.Options.nodesep}}";
	ordering="out";

    node [shape="box" style="filled" fontname="Ubuntu" fillcolor="honeydew" penwidth="1.0" margin="0.05,0.0"];
    edge [minlen="{{.Options.minlen}}"]

    {{template "cluster" .Cluster}}

    {{- range .Edges}}
    {{template "edge" .}}
    {{- end}}

	{{range .Nodes}}
	{{template "node" .}}
	{{- end}}
}
`
const tmplNode = `{{define "edge" -}}
    {{printf "%q -> %q [ %s ]" .From .To .Attrs}}
{{- end}}`

const tmplEdge = `{{define "node" -}}
    {{printf "%q [ %s ]" .ID .Attrs}}
{{- end}}`

const tmplCluster = `{{define "cluster" -}}
    {{printf "subgraph %q {" .}}
        {{printf "%s" .Attrs.Lines}}
        {{range .Nodes}}
        	{{template "node" .}}
        {{- end}}
        {{range .Clusters}}
        	{{template "cluster" .}}
        {{- end}}
    {{println "}" }}
{{- end}}`

type dotGraph struct {
	Title   string
	Minlen  uint
	Attrs   dotAttrs
	Cluster *dotCluster
	Nodes   []*dotNode
	Edges   []*dotEdge
	Options map[string]string
}

type dotCluster struct {
	ID       string
	Clusters map[string]*dotCluster
	Nodes    []*dotNode
	Attrs    dotAttrs
}

type dotNode struct {
	ID    string
	Attrs dotAttrs
}

type dotEdge struct {
	From  *dotNode
	To    *dotNode
	Attrs dotAttrs
}

type dotAttrs map[string]string

func NewDotCluster(id string) *dotCluster {
	return &dotCluster{
		ID:       id,
		Clusters: make(map[string]*dotCluster),
		Attrs:    make(dotAttrs),
	}
}

func (c *dotCluster) String() string {
	return fmt.Sprintf("cluster_%s", c.ID)
}
func (n *dotNode) String() string {
	return n.ID
}

func (p dotAttrs) List() []string {
	l := []string{}
	for k, v := range p {
		l = append(l, fmt.Sprintf("%s=%q", k, v))
	}
	return l
}

func (p dotAttrs) String() string {
	return strings.Join(p.List(), " ")
}

func (p dotAttrs) Lines() string {
	return fmt.Sprintf("%s;", strings.Join(p.List(), ";\n"))
}

func WriteDot(w io.Writer, g *dotGraph) error {
	t := template.New("dot")
	for _, s := range []string{tmplCluster, tmplNode, tmplEdge, tmplGraph} {
		if _, err := t.Parse(s); err != nil {
			return err
		}
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, g); err != nil {
		return err
	}
	_, err := buf.WriteTo(w)
	return err
}
