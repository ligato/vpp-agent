package utils

import (
	"bytes"
	"fmt"
	"strconv"
)

type printLine struct {
	line    string
	lnLevel int
	subtree []printLine
}

type lineBuf []printLine

// TreeWriter is an implementation of the TreePrinter interface.
type TreeWriter struct {
	writeBuf []byte
	level    int
	lineBuf  []printLine

	spaces     int
	firstDash  string
	middleDash string
	lastDash   string
}

// NewTreeWriter returns a reference to a newly created TreeWriter
// instance. Parameters passed into this function determine the visual
// appearance of the tree in which data is printed. A typical usage
// would be:
//   p := NewTreeWriter(1, "├─", "│ ", "└─")
func NewTreeWriter(spaces int, first string, middle string, last string) *TreeWriter {
	return &TreeWriter{
		writeBuf:   []byte{},
		lineBuf:    []printLine{},
		spaces:     spaces,
		firstDash:  first,
		middleDash: middle,
		lastDash:   last,
	}
}

// FlushTree takes the contant of the finalize buffer formats it
// into a tree and prints it out to stdout.
func (p *TreeWriter) FlushTree() {

	p.lineBuf = createPrintLineBuf(p.writeBuf)
	//for i, lbl := range p.lineBuf {
	//	fmt.Printf("%d: Level %d, Line '%s'\n", i, lbl.lnLevel, lbl.line)
	//}

	tree, _ := createTree(1, p.lineBuf)
	stack := &pfxStack{
		entries:    []pfxStackEntry{},
		spaces:     p.spaces,
		firstDash:  p.firstDash,
		middleDash: p.middleDash,
		lastDash:   p.lastDash,
	}
	stack.push()
	p.renderSubtree(tree, stack)
	p.writeBuf = []byte{}
}

// createPrintLineBuf creates a new buffer of printLine structs
// that are then used to create a printLine tree which is used to
// render the tree. The function translates the content of a raw
// write buffer into a flat buffer of printLines.
//
// The function expects that each line in the raw write buffer contains
// printLine level information - each line in the write buffer is
// expected to have the format '<level>^@<content-of-the-line>, where
// '^@' is the separator.
func createPrintLineBuf(byteBuf []byte) []printLine {
	lines := bytes.Split(bytes.TrimSpace(byteBuf), []byte{10})

	printLineBuf := make([]printLine, 0, len(lines)+1)
	for _, line := range lines {
		lbl := printLine{}
		if len(line) == 0 {
			lbl.line = string(line)
		} else {
			aux := bytes.Split(line, []byte{'^', '@'})
			lbl.lnLevel, _ = strconv.Atoi(string(aux[0]))
			lbl.line = string(bytes.TrimSpace(aux[1]))
		}
		printLineBuf = append(printLineBuf, lbl)
	}
	for i, lbl := range printLineBuf {
		if len(lbl.line) == 0 {
			lbl.lnLevel = printLineBuf[i+1].lnLevel
			printLineBuf[i] = lbl
		}
	}
	return printLineBuf
}

// renderSubtree is used to recursively render the tree
func (p *TreeWriter) renderSubtree(tree []printLine, stack *pfxStack) {
	for i, pl := range tree {
		if i == len(tree)-1 {
			stack.setLast()
		}

		var pp string
		if pl.line == "" {
			pp = stack.setTopPfxStackEntry(stack.getPreamble(stack.middleDash))
		} else {
			pp = stack.getTopPfxStackEntry()
		}
		//fmt.Printf("%2d of %2d: level %d, Line: '%s %s'\n",
		// 		i, len(tree), pl.lnLevel, stack.getPrefix(), pl.line)
		fmt.Printf("%s %s\n", stack.getPrefix(), pl.line)
		stack.setTopPfxStackEntry(pp)

		if len(pl.subtree) > 0 {
			stack.push()
			p.renderSubtree(pl.subtree, stack)
			stack.pop()
		}
	}
}

// createTree creates a tree of printLine structs from a flat printLine
// buffer (typically created in createPrintLineBuf()).
func createTree(curLevel int, lineBuf []printLine) ([]printLine, int) {
	//fmt.Printf("--> Enter createTree: curLevel %d, lineBufLen %d, line[0]: %s\n",
	// 	curLevel, len(lineBuf), lineBuf[0].line)
	res := []printLine{}
	processed := 0
	lb := lineBuf

Loop:
	for len(lb) > 0 {
		// fmt.Printf("   lb[0]: lnLevel %d, line '%s'\n", lb[0].lnLevel, lb[0].line)
		if lb[0].lnLevel < curLevel {
			break Loop
		} else if lb[0].lnLevel == curLevel {
			res = append(res, lb[0])
			processed++
			lb = lb[1:]
		} else {
			subtree, p := createTree(lb[0].lnLevel, lb)
			res[len(res)-1].subtree = subtree
			processed += p
			lb = lb[p:]
		}
	}
	//fmt.Printf("<-- Return createTree: curLevel %d, lineBufLen %d, line[0]: %s\n",
	//	curLevel, len(lineBuf), lineBuf[0].line)
	return res, processed
}

// Write is an override of io.Write - it just collects the data
// to be written in a holding buffer for later printing in the
// FlushTable() function.
func (p *TreeWriter) Write(b []byte) (n int, err error) {
	// fmt.Printf("'%s'", b)
	p.writeBuf = append(p.writeBuf[:], b[:]...)
	return len(b), nil
}
