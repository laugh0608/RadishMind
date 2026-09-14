package workspacepolicy

import (
	"slices"
	"testing"
)

func TestPermissionValidationAndNormalization(t *testing.T) {
	tests := []struct {
		name       string
		input      []string
		want       []string
		validGrant bool
	}{
		{name: "nil"},
		{name: "empty", input: []string{}},
		{name: "blank", input: []string{" \t"}},
		{name: "unknown", input: []string{"applications:read", "unknown:read"}},
		{name: "case sensitive", input: []string{"Applications:read"}},
		{name: "trim", input: []string{" runs:read\t", "applications:read"}, want: []string{"runs:read", "applications:read"}, validGrant: true},
		{name: "duplicate", input: []string{"runs:read", "runs:read"}, want: []string{"runs:read"}},
		{name: "trimmed duplicate", input: []string{" runs:read", "applications:read", "runs:read "}, want: []string{"runs:read", "applications:read"}},
		{name: "invalid after duplicate", input: []string{"runs:read", "runs:read", ""}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			original := slices.Clone(test.input)
			got, valid := NormalizeRequiredPermissions(test.input)
			if valid != (test.want != nil) || !slices.Equal(got, test.want) {
				t.Fatalf("normalization: got=%v valid=%v want=%v", got, valid, test.want)
			}
			if ValidPermissionGrants(test.input) != test.validGrant {
				t.Fatalf("grants validation: want=%v", test.validGrant)
			}
			if !slices.Equal(test.input, original) {
				t.Fatal("validation mutated the caller's input")
			}
			if len(got) > 0 {
				got[0] = "mutated"
				if !slices.Equal(test.input, original) {
					t.Fatal("normalized permissions alias the caller's input")
				}
			}
		})
	}
}

func TestPermissionNamesAreSortedIndependentCopies(t *testing.T) {
	names := PermissionNames()
	if len(names) != 68 || !slices.IsSorted(names) || !ValidPermissionGrants(names) {
		t.Fatalf("permission directory drifted: %v", names)
	}
	original := slices.Clone(names)
	names[0] = "unknown:write"
	if !slices.Equal(PermissionNames(), original) {
		t.Fatal("caller mutated canonical permission directory")
	}
}

func TestManagementPermissionMatching(t *testing.T) {
	management := []string{PermissionMembersRead, PermissionMembershipsWrite, PermissionRolesRead, PermissionRolesAssign}
	for _, permission := range management {
		if !ContainsManagementPermission([]string{permission}) || HasManagementPermissions([]string{permission}) {
			t.Fatalf("single management permission matched incorrectly: %s", permission)
		}
		if ContainsManagementPermission([]string{" " + permission}) {
			t.Fatal("management matching must preserve exact grant semantics")
		}
	}
	if !HasManagementPermissions(management) || !ContainsManagementPermission(management) {
		t.Fatal("complete management grants rejected")
	}
	for i := range management {
		partial := slices.Delete(slices.Clone(management), i, i+1)
		if HasManagementPermissions(partial) {
			t.Fatalf("missing management permission accepted: %s", management[i])
		}
	}
	for _, grants := range [][]string{nil, {}, {""}, {"applications:read"}, {"unknown:write"}} {
		if ContainsManagementPermission(grants) || HasManagementPermissions(grants) {
			t.Fatalf("non-management grants accepted: %v", grants)
		}
	}
}
