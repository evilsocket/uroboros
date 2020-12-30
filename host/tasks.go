package host

import (
	"bytes"
	"fmt"
	"github.com/evilsocket/islazy/fs"
	"github.com/evilsocket/islazy/str"
	"io/ioutil"
	"path"
	"strconv"
	"strings"
)

type Task struct {
	ID      int
	IDStr   string
	Path    string
	Comm    string
	CmdLine []string
	Stack   Stack
}

func (c Task) String() string {
	return fmt.Sprintf("(%s) %s", c.IDStr, c.Comm)
}

func parseProcessTasks(pid int) ([]Task, error) {
	tasks := make([]Task, 0)
	tasksPath := fmt.Sprintf("%s/%d/task/", ProcFS, pid)

	err := fs.Glob(tasksPath, "*", func(taskPath string) error {
		taskIDStr := path.Base(taskPath)
		if taskID, e := strconv.Atoi(taskIDStr); e == nil {
			task := Task{
				ID:    taskID,
				IDStr: taskIDStr,
				Path:  taskPath,
			}

			cmdLinePath := fmt.Sprintf("%s/cmdline", taskPath)
			if data, e := ioutil.ReadFile(cmdLinePath); e != nil {
				return e
			} else {
				task.CmdLine = strings.Split(string(bytes.TrimRight(data, "\x00")), string(byte(0)))
			}

			commLinePath := fmt.Sprintf("%s/comm", taskPath)
			if data, e := ioutil.ReadFile(commLinePath); e != nil {
				return e
			} else {
				task.Comm = str.Trim(string(data))
			}

			if task.Stack, e = parseStack(taskPath); e != nil {
				return e
			}

			tasks = append(tasks, task)
		}

		return nil
	})

	return tasks, err
}
