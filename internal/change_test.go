package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var c1 = Change{
	RegistryName: "my-registry",
	ResourceName: "my-resource",
	OldVersion:   "1.0.0",
	NewVersion:   "2.0.0",
	OldValue:     "my-image:1.0.0",
	NewValue:     "my-image:2.0.0",
	File:         "folder/file.yaml",
	LineNum:      3,
}

var c2 = Change{
	RegistryName: "my-registry2",
	ResourceName: "my-resource2",
	OldVersion:   "3.0.0",
	NewVersion:   "4.0.0",
	OldValue:     "my-image2:3.0.0",
	NewValue:     "my-image2:4.0.0",
	File:         "folder/file2.yaml",
	LineNum:      10,
}

func TestChangeIdentifier(t *testing.T) {
	assert.Equal(t, "folder/file.yaml#3#my-image:2.0.0", c1.Identifier())
	assert.Equal(t, "folder/file2.yaml#10#my-image2:4.0.0", c2.Identifier())
}

func TestChangesTitle(t *testing.T) {
	assert.Equal(t, "Update my-registry/my-resource:2.0.0", Changes{c1}.Title())
	assert.Equal(t, "Update my-registry/my-resource:2.0.0, my-registry2/my-resource2:4.0.0", Changes{c1, c2}.Title())
}

func TestChangesBranch(t *testing.T) {
	assert.Equal(t, "git-ops-update/my-registry-my-resource-2.0.0/f947389f23a7d6f7", Changes{c1}.Branch("git-ops-update"))
	assert.Equal(t, "git-ops-update/my-registry-my-resource-2.0.0-my-registry2-my-resource2-4.0.0/cf0b28c1d50d0788", Changes{c1, c2}.Branch("git-ops-update"))
}
