package internal

import "fmt"

type Action interface {
	Apply(dir string, changes Changes) error
	AlreadyApplied(dir string, changes Changes) bool
}

type PushAction struct {
	git GitProvider
}

func (a PushAction) Apply(dir string, changes Changes) error {
	return a.git.Push(dir, changes)
}
func (a PushAction) AlreadyApplied(dir string, changes Changes) bool {
	return false
}

type RequestAction struct {
	git GitProvider
}

func (a RequestAction) Apply(dir string, changes Changes) error {
	return a.git.Request(dir, changes)
}
func (a RequestAction) AlreadyApplied(dir string, changes Changes) bool {
	return a.git.AlreadyRequested(dir, changes)
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
