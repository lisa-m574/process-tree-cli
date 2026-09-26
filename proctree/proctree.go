// Package proctree builds and prints process trees from a flat list of
// (pid, ppid, name) records. The tree-building logic here has no
// dependency on how the records were obtained, so it works the same
// whether they came from /proc or from a test fixture.
package proctree

import (
	"fmt"
	"io"
	"sort"
	"strings"
)

// Process is one entry in a process table: a pid, its parent pid, and
// whatever name the kernel or caller associates with it.
type Process struct {
	PID  int
	PPID int
	Name string
}

// Node is a Process placed in a tree, with pointers to its children.
type Node struct {
	Process
	Children []*Node
}

// BuildTree links a flat list of processes into parent/child trees and
// returns the roots. A process is treated as a root if its PPID does not
// appear in the input (its parent already exited, or it's pid 1) or if
// its PPID equals its own PID, which happens for a handful of kernel
// threads and would otherwise create a cycle.
func BuildTree(procs []Process) []*Node {
	byPID := make(map[int]*Node, len(procs))
	for _, p := range procs {
		byPID[p.PID] = &Node{Process: p}
	}

	var roots []*Node
	for _, p := range procs {
		n := byPID[p.PID]
		parent, ok := byPID[p.PPID]
		if !ok || p.PPID == p.PID {
			roots = append(roots, n)
			continue
		}
		parent.Children = append(parent.Children, n)
	}

	sortByPID(roots)
	for _, n := range byPID {
		sortByPID(n.Children)
	}
	return roots
}

func sortByPID(nodes []*Node) {
	sort.Slice(nodes, func(i, j int) bool { return nodes[i].PID < nodes[j].PID })
}

// Find walks the tree looking for a node with the given pid, returning
// nil if none is found.
func Find(nodes []*Node, pid int) *Node {
	for _, n := range nodes {
		if n.PID == pid {
			return n
		}
		if found := Find(n.Children, pid); found != nil {
			return found
		}
	}
	return nil
}

// FilterByName returns a new forest containing only nodes whose Name
// contains substr, plus any ancestor needed to keep a matching node
// reachable from a root. The input tree is not modified. A node with no
// matching descendants and a non-matching name is dropped, so filtering
// for e.g. "sh" still shows the chain of parents leading to a matching
// bash or ssh process, not just that process in isolation.
func FilterByName(nodes []*Node, substr string) []*Node {
	var out []*Node
	for _, n := range nodes {
		children := FilterByName(n.Children, substr)
		if len(children) == 0 && !strings.Contains(n.Name, substr) {
			continue
		}
		out = append(out, &Node{Process: n.Process, Children: children})
	}
	return out
}

// Fprint writes the tree in a pstree-like format, e.g.:
//
//	init (1)
//	├── sshd (842)
//	│   └── bash (1011)
//	└── cron (901)
func Fprint(w io.Writer, roots []*Node) {
	for _, n := range roots {
		fmt.Fprintf(w, "%s (%d)\n", n.Name, n.PID)
		printChildren(w, n.Children, "")
	}
}

func printChildren(w io.Writer, nodes []*Node, prefix string) {
	for i, n := range nodes {
		last := i == len(nodes)-1

		connector := "├── "
		nextPrefix := prefix + "│   "
		if last {
			connector = "└── "
			nextPrefix = prefix + "    "
		}

		fmt.Fprintf(w, "%s%s%s (%d)\n", prefix, connector, n.Name, n.PID)
		printChildren(w, n.Children, nextPrefix)
	}
}
