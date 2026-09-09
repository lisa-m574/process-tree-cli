# process-tree-cli

`ps -ef` gives you a flat list of processes. Figuring out who spawned
what, which shell owns a runaway build, or why a daemon has forty
orphaned children under it means reconstructing the tree in your head
from PID/PPID columns. This is a small Go library that does that
reconstruction for you, plus a CLI that prints the result.

## Library

The core of the package doesn't care where the process list came from.
`BuildTree` takes a flat `[]proctree.Process` and links it into a
forest of `*proctree.Node`:

```go
procs := []proctree.Process{
    {PID: 1, PPID: 0, Name: "init"},
    {PID: 842, PPID: 1, Name: "sshd"},
    {PID: 1011, PPID: 842, Name: "bash"},
    {PID: 901, PPID: 1, Name: "cron"},
}

roots := proctree.BuildTree(procs)
proctree.Fprint(os.Stdout, roots)
```

```
init (1)
├── sshd (842)
│   └── bash (1011)
└── cron (901)
```

On Linux, `proctree.ReadProcesses()` builds that `[]Process` for you by
reading `/proc/[pid]/stat` for every running process. On other
platforms it returns an error — the tree-building and printing code
still works fine, you just have to supply the process list yourself.

`proctree.Find(roots, pid)` walks the tree and returns the node for a
given pid, or nil.

## CLI

```
go run ./cmd/proctree
```

prints the whole tree for the current machine. To zoom in on one
subtree:

```
go run ./cmd/proctree -pid 1011
```

## Status

First pass. Linux only for live process reading, no filtering by name,
no JSON output yet. See the library section above if you want the tree
logic without any of that.
