package internal

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type Change struct {
	RegistryName string
	ResourceName string
	OldVersion   string
	NewVersion   string
	File         string
	FileFormat   FileFormat
	LineNum      int
	OldValue     string
	NewValue     string
	Comments     string
}

type Changes []Change

const gitHubMaxPullRequestTitleLength = 256

func (cs Changes) GroupHash() string {
	temp := []byte{}
	for _, c := range cs {
		cHash := sha256.Sum256([]byte(fmt.Sprintf("%s#%d", c.File, c.LineNum)))
		temp = append(temp, cHash[:]...)
	}
	hash := sha256.Sum256(temp)
	return capString(hex.EncodeToString(hash[:]), 16)
}

func (cs Changes) Hash() string {
	temp := []byte{}
	for _, c := range cs {
		cHash := sha256.Sum256([]byte(fmt.Sprintf("%s#%d#%s", c.File, c.LineNum, c.NewValue)))
		temp = append(temp, cHash[:]...)
	}
	hash := sha256.Sum256(temp)
	return capString(hex.EncodeToString(hash[:]), 16)
}

func (cs Changes) Branch(prefix string) string {
	updates := []string{}
	for _, change := range cs {
		updates = append(updates, fmt.Sprintf("%s/%s:%s", change.RegistryName, change.ResourceName, change.NewVersion))
	}

	return fmt.Sprintf(
		"%s/%s/%s/%s",
		cs.BranchFindPrefix(prefix),
		capString(dashCased(strings.Join(updates, "-")), 128),
		cs.GroupHash(),
		cs.Hash())
}

func (cs Changes) BranchFindPrefix(prefix string) string {
	return prefix
}

func (c Change) Message() string {
	return fmt.Sprintf("Update %s:%d from %s to %s", c.File, c.LineNum, c.OldValue, c.NewValue)
}

func (cs Changes) Title() string {
	updates := []string{}
	for _, change := range cs {
		updates = append(updates, fmt.Sprintf("%s/%s:%s", change.RegistryName, change.ResourceName, change.NewVersion))
	}
	result := fmt.Sprintf("Update %s", strings.Join(updates, ", "))
	if len(result) > gitHubMaxPullRequestTitleLength {
		return result[0:gitHubMaxPullRequestTitleLength]
	}
	return result
}

func (cs Changes) Message() string {
	changeCommments := []string{}
	for _, c := range cs {
		changeCommments = append(changeCommments, strings.Trim(c.Message()+"\n"+c.Comments, "\n "))
	}
	return strings.Join(changeCommments, "\n\n---\n\n")
}

func (c Change) Push(dir string, fileHooks ...func(file string) error) error {
	file := filepath.Join(dir, c.File)
	bytes, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	lines := strings.Split(string(bytes), "\n")

	err = c.FileFormat.WriteValue(lines, c.LineNum, c.NewValue)
	if err != nil {
		return err
	}

	err = os.WriteFile(file, []byte(strings.Join(lines, "\n")), 0o664)
	if err != nil {
		return err
	}
	for _, hook := range fileHooks {
		err := hook(c.File)
		if err != nil {
			return err
		}
	}
	return nil
}

func (cs Changes) Push(dir string, fileHooks ...func(file string) error) error {
	for _, c := range cs {
		err := c.Push(dir, fileHooks...)
		if err != nil {
			return err
		}
	}
	return nil
}

var dashCaseReplaceRegex = regexp.MustCompile("[^a-z0-9.]+")

func dashCased(str string) string {
	lower := strings.ToLower(str)
	dashed := dashCaseReplaceRegex.ReplaceAllString(lower, "-")
	trimmed := strings.Trim(dashed, "-")
	return trimmed
}

func capString(str string, max int) string {
	if len(str) > max {
		return str[:max]
	}
	return str
}
