// Package proctree builds and prints process trees from a flat list of
// (pid, ppid, name) records. The tree-building logic here has no
// dependency on how the records were obtained, so it works the same
// whether they came from /proc or from a test fixture.
package proctree

import (
	"fmt"
	"io"
	"sort"
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
