//go:build linux

package proctree

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// ReadProcesses lists every process currently visible under /proc and
// returns it as a flat slice, ready for BuildTree. Processes that exit
// while we're scanning are silently skipped rather than treated as an
// error, since that race is normal on a live system.
func ReadProcesses() ([]Process, error) {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return nil, fmt.Errorf("proctree: reading /proc: %w", err)
	}

	var procs []Process
	for _, e := range entries {
		pid, err := strconv.Atoi(e.Name())
		if err != nil {
			continue // not a pid directory (self, net, etc.)
		}

		data, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", pid))
		if err != nil {
			continue // process exited between ReadDir and here
		}

		p, err := parseStat(pid, string(data))
		if err != nil {
			continue
		}
		procs = append(procs, p)
	}
	return procs, nil
}

// parseStat pulls pid, comm, and ppid out of a /proc/[pid]/stat line.
// The format is "pid (comm) state ppid ...", and comm is parsed between
// the first '(' and the last ')' because the command name itself can
// contain spaces or parentheses.
func parseStat(pid int, line string) (Process, error) {
	open := strings.IndexByte(line, '(')
	closeParen := strings.LastIndexByte(line, ')')
	if open < 0 || closeParen < open {
		return Process{}, fmt.Errorf("malformed stat line for pid %d", pid)
	}
	name := line[open+1 : closeParen]

	fields := strings.Fields(line[closeParen+1:])
	if len(fields) < 2 {
		return Process{}, fmt.Errorf("malformed stat line for pid %d", pid)
	}
	ppid, err := strconv.Atoi(fields[1])
	if err != nil {
		return Process{}, fmt.Errorf("malformed ppid for pid %d: %w", pid, err)
	}

	return Process{PID: pid, PPID: ppid, Name: name}, nil
}
