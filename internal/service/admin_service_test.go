package service

import "testing"

func TestValidateIDs(t *testing.T) {
	tests := []struct {
		name string
		ids  []int64
		want bool
	}{
		{name: "empty is allowed", ids: nil, want: true},
		{name: "positive unique ids", ids: []int64{1, 2, 3}, want: true},
		{name: "zero is rejected", ids: []int64{1, 0}, want: false},
		{name: "negative is rejected", ids: []int64{-1}, want: false},
		{name: "duplicates are rejected", ids: []int64{2, 2}, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := validateIDs(tt.ids) == nil; got != tt.want {
				t.Fatalf("validateIDs(%v) valid = %v, want %v", tt.ids, got, tt.want)
			}
		})
	}
}

func TestValidatePassword(t *testing.T) {
	if err := validatePassword("12345"); err == nil {
		t.Fatal("expected short password to be rejected")
	}
	if err := validatePassword("123456"); err != nil {
		t.Fatalf("expected six-character password to be accepted: %v", err)
	}
	if err := validatePassword(string(make([]byte, 73))); err == nil {
		t.Fatal("expected long password to be rejected")
	}
}

func TestValidateRoleCode(t *testing.T) {
	for _, code := range []string{"", "role code", string(make([]byte, 65))} {
		if err := validateRoleCode(code); err == nil {
			t.Fatalf("expected role code %q to be rejected", code)
		}
	}
	if err := validateRoleCode("content_editor"); err != nil {
		t.Fatalf("expected role code to be accepted: %v", err)
	}
}
