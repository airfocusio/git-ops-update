package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var c1 = Change{
	RegistryName:   "my-registry",
	ResourceName:   "my-resource",
	OldVersion:     "1.0.0",
	NewVersion:     "2.0.0",
	OldValue:       "my-image:1.0.0",
	NewValue:       "my-image:2.0.0",
	File:           "folder/file.yaml",
	LineNum:        3,
	RenderComments: func() (string, string) { return "Line", "Footer1" },
}

var c2 = Change{
	RegistryName:   "my-registry2",
	ResourceName:   "my-resource2",
	OldVersion:     "3.0.0",
	NewVersion:     "4.0.0",
	OldValue:       "my-image2:3.0.0",
	NewValue:       "my-image2:4.0.0",
	File:           "folder/file2.yaml",
	LineNum:        10,
	RenderComments: func() (string, string) { return "Multi\nLine", "Footer2" },
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
	assert.Equal(t, "cdb34bd928617494", ChangeSet{Changes: []Change{c1}}.GroupHash())
	assert.Equal(t, "79baa33623f66cab", ChangeSet{Changes: []Change{c2}}.GroupHash())
	assert.Equal(t, "7828b1505ed039e7", ChangeSet{Changes: []Change{c1, c2}}.GroupHash())
	assert.Equal(t, "301e5175dc9cc3f0", ChangeSet{Group: "c1c2", Changes: []Change{c1, c2}}.GroupHash())
}

func TestChangesHash(t *testing.T) {
	assert.Equal(t, "f947389f23a7d6f7", ChangeSet{Changes: []Change{c1}}.Hash())
	assert.Equal(t, "c3de406196d88bed", ChangeSet{Changes: []Change{c2}}.Hash())
	assert.Equal(t, "cf0b28c1d50d0788", ChangeSet{Changes: []Change{c1, c2}}.Hash())
	assert.Equal(t, "cf0b28c1d50d0788", ChangeSet{Group: "c1c2", Changes: []Change{c1, c2}}.Hash())
}

func TestChangesTitle(t *testing.T) {
	assert.Equal(t, "Update my-registry/my-resource:2.0.0", ChangeSet{Changes: []Change{c1}}.Title())
	assert.Equal(t, "Update my-registry/my-resource:2.0.0, my-registry2/my-resource2:4.0.0", ChangeSet{Changes: []Change{c1, c2}}.Title())
}

func TestChangesMessage(t *testing.T) {
	m1, m1Full := ChangeSet{Changes: []Change{c1}}.Message()
	assert.Equal(t, "Update folder/file.yaml:3 from my-image:1.0.0 to my-image:2.0.0\nLine", m1)
	assert.Equal(t, "Update folder/file.yaml:3 from my-image:1.0.0 to my-image:2.0.0\nLine\n\nFooter1", m1Full)

	m2, m2Full := ChangeSet{Changes: []Change{c1, c2}}.Message()
	assert.Equal(t, "Update folder/file.yaml:3 from my-image:1.0.0 to my-image:2.0.0\nLine\n\n---\n\nUpdate folder/file2.yaml:10 from my-image2:3.0.0 to my-image2:4.0.0\nMulti\nLine", m2)
	assert.Equal(t, "Update folder/file.yaml:3 from my-image:1.0.0 to my-image:2.0.0\nLine\n\n---\n\nUpdate folder/file2.yaml:10 from my-image2:3.0.0 to my-image2:4.0.0\nMulti\nLine\n\nFooter1\nFooter2", m2Full)

	m3, m3Full := ChangeSet{Changes: []Change{c3}}.Message()
	assert.Equal(t, "Update folder/file3.yaml:16 from my-image3:5.0.0 to my-image3:6.0.0", m3)
	assert.Equal(t, "Update folder/file3.yaml:16 from my-image3:5.0.0 to my-image3:6.0.0", m3Full)
}

func TestChangesBranch(t *testing.T) {
	assert.Equal(t, "git-ops-update/my-registry-my-resource-2.0.0/cdb34bd928617494/f947389f23a7d6f7", ChangeSet{Changes: []Change{c1}}.Branch("git-ops-update"))
	assert.Equal(t, "git-ops-update/my-registry-my-resource-2.0.0-my-registry2-my-resource2-4.0.0/7828b1505ed039e7/cf0b28c1d50d0788", ChangeSet{Changes: []Change{c1, c2}}.Branch("git-ops-update"))
}
