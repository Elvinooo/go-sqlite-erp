package auth

import "testing"

func TestValidatePasswordPolicy(t *testing.T) {
	tests := []struct {
		name      string
		password  string
		wantError bool
	}{
		{name: "valid", password: "Abc12345", wantError: false},
		{name: "too short", password: "Abc123", wantError: true},
		{name: "missing digit", password: "Password", wantError: true},
		{name: "missing letter", password: "12345678", wantError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePasswordPolicy(tt.password)
			if (err != nil) != tt.wantError {
				t.Fatalf("validatePasswordPolicy(%q) error = %v, wantError %v", tt.password, err, tt.wantError)
			}
		})
	}
}
