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
	Comments:     "Line",
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
	Comments:     "Multi\nLine\n",
}

var c3 = Change{
	RegistryName: "my-registry3",
	ResourceName: "my-resource3",
	OldVersion:   "5.0.0",
	NewVersion:   "6.0.0",
	OldValue:     "my-image3:5.0.0",
	NewValue:     "my-image3:6.0.0",
	File:         "folder/file3.yaml",
	LineNum:      16,
}

func TestChangesGroupHash(t *testing.T) {
	assert.Equal(t, "cdb34bd928617494", Changes{c1}.GroupHash())
	assert.Equal(t, "79baa33623f66cab", Changes{c2}.GroupHash())
	assert.Equal(t, "7828b1505ed039e7", Changes{c1, c2}.GroupHash())
}

func TestChangesHash(t *testing.T) {
	assert.Equal(t, "f947389f23a7d6f7", Changes{c1}.Hash())
	assert.Equal(t, "c3de406196d88bed", Changes{c2}.Hash())
	assert.Equal(t, "cf0b28c1d50d0788", Changes{c1, c2}.Hash())
}

func TestChangesTitle(t *testing.T) {
	assert.Equal(t, "Update my-registry/my-resource:2.0.0", Changes{c1}.Title())
	assert.Equal(t, "Update my-registry/my-resource:2.0.0, my-registry2/my-resource2:4.0.0", Changes{c1, c2}.Title())
}

func TestChangesMessage(t *testing.T) {
	assert.Equal(t, "Update folder/file.yaml:3 from my-image:1.0.0 to my-image:2.0.0\nLine", Changes{c1}.Message())
	assert.Equal(t, "Update folder/file.yaml:3 from my-image:1.0.0 to my-image:2.0.0\nLine\n\n---\n\nUpdate folder/file2.yaml:10 from my-image2:3.0.0 to my-image2:4.0.0\nMulti\nLine", Changes{c1, c2}.Message())

	assert.Equal(t, "Update folder/file3.yaml:16 from my-image3:5.0.0 to my-image3:6.0.0", Changes{c3}.Message())
}

func TestChangesBranch(t *testing.T) {
	assert.Equal(t, "git-ops-update/my-registry-my-resource-2.0.0/cdb34bd928617494/f947389f23a7d6f7", Changes{c1}.Branch("git-ops-update"))
	assert.Equal(t, "git-ops-update/my-registry-my-resource-2.0.0-my-registry2-my-resource2-4.0.0/7828b1505ed039e7/cf0b28c1d50d0788", Changes{c1, c2}.Branch("git-ops-update"))
}
