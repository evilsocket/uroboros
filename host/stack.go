package host

import (
	"bufio"
	"fmt"
	"github.com/evilsocket/islazy/str"
	"os"
	"regexp"
)

type StackEntry struct {
	Address  uint
	Function string
	Offset   uint
	Size     uint
}

type Stack []StackEntry

var stackParser = regexp.MustCompile(`(?i)\[<([a-f0-9]+)>\]\s+(.+)\+0x([a-f0-9]+)\/0x([a-f0-9]+)`)

func parseStack(taskPath string) (Stack, error) {
	filename := fmt.Sprintf("%s/stack", taskPath)
	fd, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer fd.Close()

	var stack Stack

	scanner := bufio.NewScanner(fd)
	for lineno := 0; scanner.Scan(); lineno++ {
		var entry StackEntry

		line := str.Trim(scanner.Text())
		if line == "[<0>] 0xffffffffffffffff" {
			continue
		}

		m := stackParser.FindStringSubmatch(line)
		if m == nil {
			return nil, fmt.Errorf("could not parse stack line from %s: %s", filename, line)
		}

		entry = StackEntry{
			Address:  hexToInt(m[1]),
			Function: m[2],
			Offset:   hexToInt(m[3]),
			Size:     hexToInt(m[4]),
		}
		stack = append(stack, entry)
	}

	return stack, nil
}
