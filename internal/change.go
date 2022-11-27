package internal

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/iancoleman/strcase"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type Change struct {
	RegistryName string
	ResourceName string
	OldVersion   string
	NewVersion   string
	Metadata     map[string]string
	File         string
	FileFormat   FileFormat
	LineNum      int
	OldValue     string
	NewValue     string
}

type Changes []Change

const gitHubMaxPullRequestTitleLength = 256

func (c Change) Identifier() string {
	return c.File + "#" + strconv.Itoa(c.LineNum) + "#" + c.NewValue
}

func (c Change) Hash() []byte {
	identifier := c.Identifier()
	hash := sha256.Sum256([]byte(identifier))
	return hash[:]
}

func (cs Changes) Hash() []byte {
	temp := []byte{}
	for _, c := range cs {
		temp = append(temp, c.Hash()...)
	}
	hash := sha256.Sum256(temp)
	return hash[:]
}

func (cs Changes) Branch(prefix string) string {
	updates := []string{}
	for _, change := range cs {
		updates = append(updates, fmt.Sprintf("%s/%s:%s", change.RegistryName, change.ResourceName, change.NewVersion))
	}

	return fmt.Sprintf(
		"%s/%s/%s",
		cs.BranchFindPrefix(prefix),
		capString(dashCased(strings.Join(updates, "-")), 128),
		cs.BranchFindSuffix())
}

func (cs Changes) BranchFindPrefix(prefix string) string {
	return prefix
}

func (cs Changes) BranchFindSuffix() string {
	return capString(hex.EncodeToString(cs.Hash()), 16)
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
	lines := []string{}
	for _, c := range cs {
		lines = append(lines, "* "+c.Message())
		metadataKeys := make([]string, 0)
		for k := range c.Metadata {
			metadataKeys = append(metadataKeys, k)
		}
		sort.Strings(metadataKeys)
		for _, k := range metadataKeys {
			value := c.Metadata[k]
			valueLines := strings.Split(value, "\n")
			for i := 1; i < len(valueLines); i++ {
				valueLines[i] = "        " + valueLines[i]
			}
			indentedValue := strings.Join(valueLines, "\n")
			niceKey := cases.Title(language.English).String(strcase.ToDelimited(k, ' '))
			lines = append(lines, "    * "+niceKey+": "+indentedValue)
		}
	}
	return strings.Join(lines, "\n")
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
