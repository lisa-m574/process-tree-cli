// Command proctree prints the process tree of the current machine, or
// the subtree rooted at a specific pid.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/lisa-m574/process-tree-cli/proctree"
)

func main() {
	pid := flag.Int("pid", 0, "only show the subtree rooted at this pid")
	flag.Parse()

	procs, err := proctree.ReadProcesses()
	if err != nil {
		fmt.Fprintln(os.Stderr, "proctree:", err)
		os.Exit(1)
	}

	roots := proctree.BuildTree(procs)

	if *pid != 0 {
		n := proctree.Find(roots, *pid)
		if n == nil {
			fmt.Fprintf(os.Stderr, "proctree: no process with pid %d\n", *pid)
			os.Exit(1)
		}
		roots = []*proctree.Node{n}
	}

	proctree.Fprint(os.Stdout, roots)
}
