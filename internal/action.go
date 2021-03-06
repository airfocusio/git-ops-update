package internal

import (
	"fmt"
	"os/exec"
	"strings"
)

type Action interface {
	Apply(dir string, changes Changes) error
	AlreadyApplied(dir string, changes Changes) bool
}

type PushAction struct {
	git  GitProvider
	exec []string
}

func (a PushAction) Apply(dir string, changes Changes) error {
	return a.git.Push(dir, changes, execCallbacks(dir, a.exec)...)
}
func (a PushAction) AlreadyApplied(dir string, changes Changes) bool {
	return false
}

type RequestAction struct {
	git  GitProvider
	exec []string
}

func (a RequestAction) Apply(dir string, changes Changes) error {
	return a.git.Request(dir, changes, execCallbacks(dir, a.exec)...)
}
func (a RequestAction) AlreadyApplied(dir string, changes Changes) bool {
	return a.git.AlreadyRequested(dir, changes)
}

func getAction(p GitProvider, actionName string, exec []string) (*Action, error) {
	switch actionName {
	case "":
		return nil, nil
	case "push":
		result := Action(PushAction{git: p, exec: exec})
		return &result, nil
	case "request":
		result := Action(RequestAction{git: p, exec: exec})
		return &result, nil
	default:
		return nil, fmt.Errorf("unknown action %s", actionName)
	}
}

func execCallbacks(dir string, execCmdAndArgs []string) []func() error {
	if len(execCmdAndArgs) == 0 {
		return []func() error{}
	}
	return []func() error{func() error {
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
	}}
}
