package services

import "testing"

// rule: §6.1 — the audit diff names every changed field with its before and after
func TestDiff(t *testing.T) {
	before := Snapshot{"id": 1.0, "name": "Old", "phone": "1", "gone": nil}
	after := Snapshot{"id": 1.0, "name": "New", "phone": "1", "gone": nil}
	changes := Diff(before, after)
	if len(changes) != 1 {
		t.Fatalf("changes = %v, want only name", changes)
	}
	if c := changes["name"]; c.Before != "Old" || c.After != "New" {
		t.Errorf("name change = %+v", c)
	}

	created := Diff(nil, Snapshot{"id": 2.0, "name": "Fresh", "technician_id": nil})
	if len(created) != 2 || created["id"].Before != nil || created["name"].After != "Fresh" {
		t.Errorf("create diff = %v, want id and name against null", created)
	}
	if _, present := created["technician_id"]; present {
		t.Error("a null field on a new row is reported as a change")
	}
}
