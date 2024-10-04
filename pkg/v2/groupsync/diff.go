package groupsync

import (
	"github.com/imulab/go-scim/pkg/v2/prop"
)

const (
	fieldDisplayName = "displayName"
	fieldMembers     = "members"
	fieldValue       = "value"
)

// Compare compares the two snapshots of two group resources before and after the modification and
// reports their differences in membership and whether relevant group properties (e.g. displayName)
// were changed. At least one of before and after should be non-nil. When before is nil, all
// members of the after resource are considered to have just joined; when after is nil, all members
// of the before resource are considered to have just left.
func Compare(before *prop.Resource, after *prop.Resource) *Diff {
	if before == nil && after == nil {
		panic("at least one of before and after should be non-nil")
	}

	diff := new(Diff)
	diff.propertiesChanged = propertiesChanged(before, after)

	var (
		beforeIds = map[string]struct{}{}
		afterIds  = map[string]struct{}{}
	)
	for _, t := range []struct {
		resource  *prop.Resource
		collector map[string]struct{}
	}{
		{resource: before, collector: beforeIds},
		{resource: after, collector: afterIds},
	} {
		if t.resource == nil {
			continue
		}

		members, _ := t.resource.RootProperty().ChildAtIndex(fieldMembers)
		_ = members.ForEachChild(func(index int, child prop.Property) error {
			value, _ := child.ChildAtIndex(fieldValue)
			if value != nil && !value.IsUnassigned() {
				t.collector[value.Raw().(string)] = struct{}{}
			}
			return nil
		})
	}

	for k := range beforeIds {
		if _, ok := afterIds[k]; !ok {
			diff.addLeft(k)
		}
	}
	for k := range afterIds {
		if _, ok := beforeIds[k]; !ok {
			diff.addJoined(k)
		} else {
			diff.addStayed(k)
		}
	}
	return diff
}

// propertiesChanged compare if relevant group properties (e.g. displayName) were changed,
// returning true in this case, or false otherwise.
func propertiesChanged(before *prop.Resource, after *prop.Resource) bool {
	// If group was created (before==nil) or deleted (after==nil), there's no relevant property
	// change, so we return false.
	if before == nil || after == nil {
		return false
	}

	displayNameBefore, _ := before.RootProperty().ChildAtIndex(fieldDisplayName)
	displayNameAfter, _ := after.RootProperty().ChildAtIndex(fieldDisplayName)

	if displayNameBefore != nil && displayNameAfter != nil {
		return displayNameBefore.Raw().(string) != displayNameAfter.Raw().(string)
	}

	return displayNameBefore != displayNameAfter
}

// Diff reports the difference between the members of two group resources.
type Diff struct {
	joined map[string]struct{}
	stayed map[string]struct{}
	left   map[string]struct{}
	// propertiesChanged indicates if relevant group properties (e.g. displayName) were changed.
	// This is relevant to only trigger updates for stayed users if relevant group changes need to be
	// updated.
	propertiesChanged bool
}

func (d *Diff) addJoined(id string) {
	if d.joined == nil {
		d.joined = map[string]struct{}{}
	}
	d.joined[id] = struct{}{}
}

func (d *Diff) addStayed(id string) {
	if d.stayed == nil {
		d.stayed = map[string]struct{}{}
	}
	d.stayed[id] = struct{}{}
}

func (d *Diff) addLeft(id string) {
	if d.left == nil {
		d.left = map[string]struct{}{}
	}
	d.left[id] = struct{}{}
}

// ForEachJoined iterates all member ids that joined the group and invoke the callback.
func (d *Diff) ForEachJoined(callback func(id string)) {
	for k := range d.joined {
		callback(k)
	}
}

// ForEachStayed iterates all member ids that stayed in the group and invoke the callback.
func (d *Diff) ForEachStayed(callback func(id string)) {
	for k := range d.stayed {
		callback(k)
	}
}

// ForEachLeft iterates all member ids that left the group and invoke the callback.
func (d *Diff) ForEachLeft(callback func(id string)) {
	for k := range d.left {
		callback(k)
	}
}

// CountJoined returns the total number of new members that joined the group.
func (d *Diff) CountJoined() int {
	return len(d.joined)
}

// CountStayed returns the total number of members that stayed in the group.
func (d *Diff) CountStayed() int {
	return len(d.stayed)
}

// CountLeft returns the total number of members that just left the group.
func (d *Diff) CountLeft() int {
	return len(d.left)
}

// PropertiesChanged indicates if relevant group properties (e.g. displayName) were changed.
// This is relevant to only trigger updates for stayed users if relevant group changes need to be
// updated.
func (d *Diff) PropertiesChanged() bool {
	return d.propertiesChanged
}
