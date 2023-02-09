package internal

import (
	"fmt"
	"os/exec"
	"strings"
)

type Action interface {
	Identifier() string
	Apply(dir string, changeSet ChangeSet) error
}

var _ Action = (*PushAction)(nil)

type PushAction struct {
	git GitProvider
}

func (c PushAction) Identifier() string {
	return "push"
}

func (a PushAction) Apply(dir string, changeSet ChangeSet) error {
	execCallbacks := SliceMap(changeSet.Changes, func(c Change) func() error {
		return execCallback(dir, c.Exec)
	})
	return a.git.Push(dir, changeSet, execCallbacks...)
}

var _ Action = (*RequestAction)(nil)

type RequestAction struct {
	git GitProvider
}

func (c RequestAction) Identifier() string {
	return "request"
}

func (a RequestAction) Apply(dir string, changeSet ChangeSet) error {
	execCallbacks := SliceMap(changeSet.Changes, func(c Change) func() error {
		return execCallback(dir, c.Exec)
	})
	return a.git.Request(dir, changeSet, execCallbacks...)
}

func getAction(p GitProvider, actionName string) (*Action, error) {
	switch actionName {
	case "":
		return nil, nil
	case "push":
		result := Action(PushAction{git: p})
		return &result, nil
	case "request":
		result := Action(RequestAction{git: p})
		return &result, nil
	default:
		return nil, fmt.Errorf("unknown action %s", actionName)
	}
}

func execCallback(dir string, execCmdAndArgs []string) func() error {
	if len(execCmdAndArgs) == 0 {
		return func() error { return nil }
	}
	return func() error {
		name := execCmdAndArgs[0]
		args := execCmdAndArgs[1:]
		LogDebug("Executing %s with args [%s]", name, strings.Join(args, ", "))
		cmd := exec.Command(name, args...)
		cmd.Dir = dir
		bytes, err := cmd.CombinedOutput()
		if err != nil {
			str := string(bytes)
			str = strings.Trim(str, " \n")
			if str != "" {
				str = "\n" + str
			}
			return fmt.Errorf("executing %s failed: %w%s", name, err, str)
		}
		return nil
	}
}
