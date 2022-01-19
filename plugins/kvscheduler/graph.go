package kvscheduler

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"google.golang.org/protobuf/encoding/prototext"

	"go.ligato.io/vpp-agent/v3/pkg/graphviz"
	kvs "go.ligato.io/vpp-agent/v3/plugins/kvscheduler/api"
	"go.ligato.io/vpp-agent/v3/plugins/kvscheduler/internal/graph"
	"go.ligato.io/vpp-agent/v3/plugins/kvscheduler/internal/utils"
	"go.ligato.io/vpp-agent/v3/proto/ligato/kvscheduler"
)

const (
	minlen = 1
)

func (s *Scheduler) renderDotOutput(graphNodes []*graph.RecordedNode, txn *kvs.RecordedTxn) ([]byte, error) {
	title := fmt.Sprintf("%d keys", len(graphNodes))
	updatedKeys := utils.NewMapBasedKeySet()
	graphTimestamp := time.Now()
	if txn != nil {
		graphTimestamp = txn.Stop
		title += fmt.Sprintf(" - SeqNum: %d (%s)", txn.SeqNum, graphTimestamp.Format(time.RFC822))
		for _, op := range txn.Executed {
			updatedKeys.Add(op.Key)
		}
	} else {
		title += " - current"
	}

	cluster := NewDotCluster("nodes")
	cluster.Attrs = dotAttrs{
		"bgcolor":   "white",
		"label":     title,
		"labelloc":  "t",
		"labeljust": "c",
		"fontsize":  "14",
		"fontname":  "Arial",
		"tooltip":   "",
	}

	// TODO: how to link transaction recording inside of the main cluster title (SeqNum: %d)?
	//if txn != nil {
	//	cluster.Attrs["href"] = fmt.Sprintf(txnHistoryURL + "?seq-num=%d", txn.SeqNum)
	//}

	var (
		nodes []*dotNode
		edges []*dotEdge
	)

	nodeMap := make(map[string]*dotNode)
	edgeMap := make(map[string]*dotEdge)

	var getGraphNode = func(key string) *graph.RecordedNode {
		for _, graphNode := range graphNodes {
			if graphNode.Key == key {
				return graphNode
			}
		}
		return nil
	}

	var processGraphNode = func(graphNode *graph.RecordedNode) *dotNode {
		key := graphNode.Key
		if n, ok := nodeMap[key]; ok {
			return n
		}

		attrs := make(dotAttrs)
		attrs["penwidth"] = "1"
		attrs["fontsize"] = "9"
		attrs["width"] = "0"
		attrs["height"] = "0"
		attrs["color"] = "Black"
		attrs["style"] = "filled"
		attrs["href"] = fmt.Sprintf(keyTimelineURL+"?key=%s&amp;time=%d", key, graphTimestamp.UnixNano())

		if updatedKeys.Has(key) {
			attrs["penwidth"] = "2"
			attrs["color"] = "Gold"
		}

		c := cluster

		label := graphNode.Label
		var descriptorName string
		if descriptorFlag := graphNode.GetFlag(DescriptorFlagIndex); descriptorFlag != nil {
			descriptorName = descriptorFlag.GetValue()
		} else {
			// for missing dependencies
			if descriptor := s.registry.GetDescriptorForKey(key); descriptor != nil {
				descriptorName = descriptor.Name
				if descriptor.KeyLabel != nil {
					label = descriptor.KeyLabel(key)
				}
			}
		}

		if label != "" {
			attrs["label"] = label
		}

		if descriptorName != "" {
			attrs["fillcolor"] = "PaleGreen"

			if _, ok := c.Clusters[descriptorName]; !ok {
				c.Clusters[descriptorName] = &dotCluster{
					ID:       descriptorName,
					Clusters: make(map[string]*dotCluster),
					Attrs: dotAttrs{
						"fontsize":  "10",
						"label":     fmt.Sprintf("< %s >", descriptorName),
						"style":     "filled",
						"fillcolor": "#e6ecfa",
						"pad":       "0.015",
						"margin":    "4",
					},
				}
			}
			c = c.Clusters[descriptorName]
		}

		var (
			dashedStyle bool
			valueState  kvscheduler.ValueState
		)
		isDerived := graphNode.GetFlag(DerivedFlagIndex) != nil
		stateFlag := graphNode.GetFlag(ValueStateFlagIndex)
		if stateFlag != nil {
			valueState = stateFlag.(*ValueStateFlag).valueState
		}

		// set colors
		switch valueState {
		case kvscheduler.ValueState_NONEXISTENT:
			attrs["fontcolor"] = "White"
			attrs["fillcolor"] = "Black"
		case kvscheduler.ValueState_MISSING:
			attrs["fillcolor"] = "Dimgray"
			dashedStyle = true
		case kvscheduler.ValueState_UNIMPLEMENTED:
			attrs["fillcolor"] = "Darkkhaki"
			dashedStyle = true
		case kvscheduler.ValueState_REMOVED:
			attrs["fontcolor"] = "White"
			attrs["fillcolor"] = "Black"
			dashedStyle = true
		// case kvs.ValueState_CONFIGURED // leave default
		case kvscheduler.ValueState_OBTAINED:
			attrs["fillcolor"] = "LightCyan"
		case kvscheduler.ValueState_DISCOVERED:
			attrs["fillcolor"] = "Lime"
		case kvscheduler.ValueState_PENDING:
			dashedStyle = true
			attrs["fillcolor"] = "Pink"
		case kvscheduler.ValueState_INVALID:
			attrs["fontcolor"] = "White"
			attrs["fillcolor"] = "Maroon"
		case kvscheduler.ValueState_FAILED:
			attrs["fillcolor"] = "Orangered"
		case kvscheduler.ValueState_RETRYING:
			attrs["fillcolor"] = "Deeppink"
		}
		if isDerived && ((valueState == kvscheduler.ValueState_CONFIGURED) ||
			(valueState == kvscheduler.ValueState_OBTAINED) ||
			(valueState == kvscheduler.ValueState_DISCOVERED)) {
			attrs["fillcolor"] = "LightYellow"
			attrs["color"] = "bisque4"
		}

		// set style
		attrs["style"] = "filled"
		if isDerived {
			attrs["style"] += ",rounded"
		}
		if dashedStyle {
			attrs["style"] += ",dashed"
		}

		value := graphNode.Value
		if rec, ok := value.(*utils.RecordedProtoMessage); ok {
			value = rec.Message
		}
		attrs["tooltip"] = fmt.Sprintf("[%s] %s\n-----\n%s", valueState, key, prototext.Format(value))

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

	for _, graphNode := range graphNodes {
		n := processGraphNode(graphNode)
		targets := graphNode.Targets

		for i := targets.RelationBegin(DerivesRelation); i < len(targets); i++ {
			if targets[i].Relation != DerivesRelation {
				break
			}
			for _, dKey := range targets[i].MatchingKeys.Iterate() {
				dn := processGraphNode(getGraphNode(dKey))
				attrs := make(dotAttrs)
				attrs["color"] = "bisque4"
				attrs["arrowhead"] = "invempty"
				e := &dotEdge{
					From:  n,
					To:    dn,
					Attrs: attrs,
				}
				addEdge(e)
			}
		}

		for i := targets.RelationBegin(DependencyRelation); i < len(targets); i++ {
			target := targets[i]
			if target.Relation != DependencyRelation {
				break
			}
			type depNode struct {
				node      *dotNode
				label     string
				satisfied bool
			}
			var deps []depNode
			if target.MatchingKeys.Length() == 0 {
				var dn *dotNode
				if target.ExpectedKey != "" {
					dn = processGraphNode(&graph.RecordedNode{
						Key: target.ExpectedKey,
					})
				} else {
					dn = processGraphNode(&graph.RecordedNode{
						Key: "? " + target.Label + " ?",
					})
				}
				deps = append(deps, depNode{node: dn, label: target.Label})
			}
			for _, dKey := range target.MatchingKeys.Iterate() {
				dn := processGraphNode(getGraphNode(dKey))
				deps = append(deps, depNode{node: dn, label: target.Label, satisfied: true})
			}
			for _, d := range deps {
				attrs := make(dotAttrs)
				attrs["tooltip"] = d.label
				if !d.satisfied {
					attrs["color"] = "Red"
				}
				e := &dotEdge{
					From:  n,
					To:    d.node,
					Attrs: attrs,
				}
				addEdge(e)
			}
		}
	}

	hostname, _ := os.Hostname()
	footer := fmt.Sprintf("KVScheduler Graph - generated at %s on %s (PID: %d)",
		time.Now().Format(time.RFC1123), hostname, os.Getpid(),
	)

	dot := &dotGraph{
		Title:   footer,
		Minlen:  minlen,
		Cluster: cluster,
		Nodes:   nodes,
		Edges:   edges,
		Options: map[string]string{
			"minlen": fmt.Sprint(minlen),
		},
	}

	var buf bytes.Buffer
	if err := WriteDot(&buf, dot); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func validateDot(output []byte) ([]byte, error) {
	dot, err := graphviz.RenderDot(output)
	if err != nil {
		return nil, fmt.Errorf("rendering dot failed: %v\nRaw output:%s", err, output)
	}
	return dot, nil
}

func dotToImage(outfname string, format string, dot []byte) (string, error) {
	var img string
	if outfname == "" {
		img = filepath.Join(os.TempDir(), fmt.Sprintf("kvscheduler-graph.%s", format))
	} else {
		img = fmt.Sprintf("%s.%s", outfname, format)
	}

	err := graphviz.RenderFilename(img, format, dot)
	if err != nil {
		return "", err
	}

	return img, nil
}

const tmplGraph = `digraph kvscheduler {
    label="{{.Title}}";
	labelloc="b";
    labeljust="c";
    fontsize="10";
    rankdir="LR";
    bgcolor="lightgray";
    style="solid";
    pad="0.035";
	ranksep="0.35";
	nodesep="0.03";
    //nodesep="{{.Options.nodesep}}";
	ordering="out";
	newrank="true";
	compound="true";

    node [shape="box" style="filled" color="black" fontname="Courier" fillcolor="honeydew"];
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
