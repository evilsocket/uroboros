package host

import (
	"bufio"
	"fmt"
	"github.com/evilsocket/islazy/fs"
	"github.com/evilsocket/islazy/str"
	"os"
	"path"
	"regexp"
)

type StackEntry struct {
	Address  uint
	Function string
	Offset   uint
	Size     uint
}

type Stack []StackEntry

type ProcessStack map[string]Stack

var stackParser = regexp.MustCompile(`(?i)\[<([a-f0-9]+)>\]\s+(.+)\+0x([a-f0-9]+)\/0x([a-f0-9]+)`)

func parseProcessStack(pid int) (ProcessStack, error) {
	stack := make(ProcessStack)
	tasksPath := fmt.Sprintf("/proc/%d/task/", pid)

	err := fs.Glob(tasksPath, "*", func(taskPath string) error {
		taskID := path.Base(taskPath)
		filename := fmt.Sprintf("%s/stack", taskPath)
		fd, err := os.Open(filename)
		if err != nil {
			return err
		}
		defer fd.Close()

		stack[taskID] = make(Stack, 0)

		scanner := bufio.NewScanner(fd)
		for lineno := 0; scanner.Scan(); lineno++ {
			var entry StackEntry

			line := str.Trim(scanner.Text())
			if line == "[<0>] 0xffffffffffffffff" {
				continue
			}

			m := stackParser.FindStringSubmatch(line)
			if m == nil {
				panic(fmt.Errorf("could not parse stack line from %s: %s", filename, line))
			}

			entry = StackEntry{
				Address:  hexToInt(m[1]),
				Function: m[2],
				Offset:   hexToInt(m[3]),
				Size:     hexToInt(m[4]),
			}
			stack[taskID] = append(stack[taskID], entry)
		}

		return nil
	})

	return stack, err
}
