package models

import "testing"

func TestRoleValid(t *testing.T) {
	for _, r := range []Role{RoleAdmin, RoleManager, RoleAnalyst, RoleViewer} {
		if !r.Valid() {
			t.Errorf("%s should be valid", r)
		}
	}
	if Role("root").Valid() {
		t.Error("root should be invalid")
	}
}

func TestRoleRankAndAtLeast(t *testing.T) {
	if !RoleAdmin.AtLeast(RoleViewer) {
		t.Error("admin should outrank viewer")
	}
	if RoleViewer.AtLeast(RoleManager) {
		t.Error("viewer should not satisfy manager")
	}
	if !RoleManager.AtLeast(RoleManager) {
		t.Error("role should satisfy itself")
	}
	if Role("bad").Rank() != 0 {
		t.Error("unknown role rank should be 0")
	}
}
