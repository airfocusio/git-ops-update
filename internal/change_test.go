package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChangeIdentifier(t *testing.T) {
	c := Change{
		RegistryName: "my-registry",
		ResourceName: "my-resource",
		OldVersion:   "1.0.0",
		NewVersion:   "2.0.0",
		OldValue:     "my-image:1.0.0",
		NewValue:     "my-image:2.0.0",
		File:         "folder/file.yaml",
		Trace:        []interface{}{"spec", 0, "image"},
	}
	assert.Equal(t, "folder/file.yaml#spec.0.image#my-image:2.0.0", c.Identifier())
}

func TestChangesBranch(t *testing.T) {
	cs := Changes{Change{
		RegistryName: "my-registry",
		ResourceName: "my-resource",
		OldVersion:   "1.0.0",
		NewVersion:   "2.0.0",
		OldValue:     "my-image:1.0.0",
		NewValue:     "my-image:2.0.0",
		File:         "folder/file.yaml",
		Trace:        []interface{}{"spec", 0, "image"},
	}}
	assert.Equal(t, "git-ops-update/b192fbd55a666b30", cs.Branch("git-ops-update"))
}
