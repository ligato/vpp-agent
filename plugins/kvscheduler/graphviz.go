package kvscheduler

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/gogo/protobuf/proto"
	kvs "github.com/ligato/vpp-agent/plugins/kvscheduler/api"
	"github.com/ligato/vpp-agent/plugins/kvscheduler/internal/graph"
	"github.com/ligato/vpp-agent/plugins/kvscheduler/internal/utils"
	"github.com/unrolled/render"
)

const (
	// txnArg allows to display graph at the time when the referenced transaction
	// has just finalized
	txnArg = "txn" // value = txn sequence number
)

type depNode struct {
	node      *dotNode
	label     string
	satisfied bool
}

func (s *Scheduler) dotGraphHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		args := req.URL.Query()
		s.txnLock.Lock()
		defer s.txnLock.Unlock()
		graphRead := s.graph.Read()
		defer graphRead.Release()

		var txn *kvs.RecordedTxn
		timestamp := time.Now()

		// parse optional *txn* argument
		if txnStr, withTxn := args[txnArg]; withTxn && len(txnStr) == 1 {
			txnSeqNum, err := strconv.ParseUint(txnStr[0], 10, 64)
			if err != nil {
				s.logError(formatter.JSON(w, http.StatusInternalServerError, errorString{err.Error()}))
				return
			}

			txn = s.GetRecordedTransaction(txnSeqNum)
			if txn == nil {
				err := errors.New("transaction with such sequence number is not recorded")
				s.logError(formatter.JSON(w, http.StatusNotFound, errorString{err.Error()}))
				return
			}
			timestamp = txn.Stop
		}

		graphSnapshot := graphRead.GetSnapshot(timestamp)
		output, err := s.renderDotOutput(graphSnapshot, txn)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if format := req.FormValue("format"); format == "dot" {
			w.Write(output)
			return
		}

		img, err := dotToImage("", "svg", output)
		if err != nil {
			http.Error(w, fmt.Sprintf("%v\n%v", err, img), http.StatusInternalServerError)
			return
		}

		s.Log.Debug("serving graph image from:", img)
		http.ServeFile(w, req, img)
	}
}

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
		"fontsize":  "15",
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
		attrs["pad"] = "0.01"
		attrs["margin"] = "0.01"
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
						"penwidth":  "0.8",
						"fontsize":  "16",
						"label":     fmt.Sprintf("< %s >", descriptorName),
						"style":     "filled",
						"fillcolor": "#e6ecfa",
					},
				}
			}
			c = c.Clusters[descriptorName]
		}

		var (
			dashedStyle bool
			valueState  kvs.ValueState
		)
		isDerived := graphNode.GetFlag(DerivedFlagIndex) != nil
		stateFlag := graphNode.GetFlag(ValueStateFlagIndex)
		if stateFlag != nil {
			valueState = stateFlag.(*ValueStateFlag).valueState
		}

		// set colors
		switch valueState {
		case kvs.ValueState_NONEXISTENT:
			attrs["fontcolor"] = "White"
			attrs["fillcolor"] = "Black"
		case kvs.ValueState_MISSING:
			attrs["fillcolor"] = "Dimgray"
			dashedStyle = true
		case kvs.ValueState_UNIMPLEMENTED:
			attrs["fillcolor"] = "Darkkhaki"
			dashedStyle = true
		case kvs.ValueState_REMOVED:
			attrs["fontcolor"] = "White"
			attrs["fillcolor"] = "Black"
			dashedStyle = true
		// case kvs.ValueState_CONFIGURED // leave default
		case kvs.ValueState_OBTAINED:
			attrs["fillcolor"] = "LightCyan"
		case kvs.ValueState_DISCOVERED:
			attrs["fillcolor"] = "Lime"
		case kvs.ValueState_PENDING:
			dashedStyle = true
			attrs["fillcolor"] = "Pink"
		case kvs.ValueState_INVALID:
			attrs["fontcolor"] = "White"
			attrs["fillcolor"] = "Maroon"
		case kvs.ValueState_FAILED:
			attrs["fillcolor"] = "Orangered"
		case kvs.ValueState_RETRYING:
			attrs["fillcolor"] = "Deeppink"
		}
		if isDerived && ((valueState == kvs.ValueState_CONFIGURED) ||
			(valueState == kvs.ValueState_OBTAINED) ||
			(valueState == kvs.ValueState_DISCOVERED)) {
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
		attrs["tooltip"] = fmt.Sprintf("[%s] %s\n-----\n%s", valueState, key, proto.MarshalTextString(value))

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

		derived := graphNode.Targets.GetTargetsForRelation(DerivesRelation)
		if derived != nil {
			for _, target := range derived.Targets {
				for _, dKey := range target.MatchingKeys.Iterate() {
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
		}

		dependencies := graphNode.Targets.GetTargetsForRelation(DependencyRelation)
		if dependencies != nil {
			var deps []depNode
			for _, target := range dependencies.Targets {
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

var (
	minlen uint = 1
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
	ranksep=.5;
	//nodesep=.1
    label="{{.Title}}";
	labelloc="b";
    labeljust="c";
    fontsize="12";
	fontname="Ubuntu"; 
    rankdir="LR";
    bgcolor="lightgray";
    style="solid";
    penwidth="1";
    pad="0.04";
    nodesep="{{.Options.nodesep}}";
	ordering="out";

    node [shape="box" style="filled" fontname="Ubuntu" fillcolor="honeydew" penwidth="1.0" margin="0.03,0.0"];
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
