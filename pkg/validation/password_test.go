package validation

import (
	"testing"
)

func TestValidatePasswordStrength(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "Valid password",
			password: "Test@1234",
			wantErr:  false,
		},
		{
			name:     "Valid password with special chars",
			password: "MyP@ssw0rd!",
			wantErr:  false,
		},
		{
			name:     "Too short",
			password: "Test@12",
			wantErr:  true,
			errMsg:   "密码长度不能少于",
		},
		{
			name:     "Too long",
			password: "ThisIsAVeryLongPasswordThatExceedsSixtyFourCharactersAndShouldBeRejected123",
			wantErr:  true,
			errMsg:   "密码长度不能超过",
		},
		{
			name:     "No uppercase",
			password: "test@1234",
			wantErr:  true,
			errMsg:   "密码必须包含至少一个大写字母",
		},
		{
			name:     "No lowercase",
			password: "TEST@1234",
			wantErr:  true,
			errMsg:   "密码必须包含至少一个小写字母",
		},
		{
			name:     "No digit",
			password: "TestPassword",
			wantErr:  true,
			errMsg:   "密码必须包含至少一个数字",
		},
		{
			name:     "Valid without special char",
			password: "TestPassword123",
			wantErr:  false, // Special char not required by default
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePasswordStrength(tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePasswordStrength() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.errMsg != "" {
				// Check if error message contains expected substring
				if !contains(err.Error(), tt.errMsg) {
					t.Errorf("ValidatePasswordStrength() error = %v, want error containing %v", err.Error(), tt.errMsg)
				}
			}
		})
	}
}

func TestContainsCommonPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		want     bool
	}{
		{
			name:     "Common password - password",
			password: "password",
			want:     true,
		},
		{
			name:     "Common password - 123456",
			password: "123456",
			want:     true,
		},
		{
			name:     "Common password - uppercase",
			password: "PASSWORD",
			want:     true,
		},
		{
			name:     "Not common password",
			password: "MySecurePass123",
			want:     false,
		},
		{
			name:     "Common password with prefix",
			password: "password123", // This should be false as it's not exact match
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ContainsCommonPassword(tt.password); got != tt.want {
				t.Errorf("ContainsCommonPassword() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestContainsSequentialChars(t *testing.T) {
	tests := []struct {
		name        string
		password    string
		maxSequence int
		want        bool
	}{
		{
			name:        "Sequential digits - 123",
			password:    "pass123word",
			maxSequence: 3,
			want:        true,
		},
		{
			name:        "Sequential digits - 456",
			password:    "test456pass",
			maxSequence: 3,
			want:        true,
		},
		{
			name:        "Sequential letters - abc",
			password:    "abcdefgh",
			maxSequence: 3,
			want:        true,
		},
		{
			name:        "Reverse sequential - 321",
			password:    "test321pass",
			maxSequence: 3,
			want:        true,
		},
		{
			name:        "No sequential",
			password:    "t3s7p9ss",
			maxSequence: 3,
			want:        false,
		},
		{
			name:        "Sequential too short",
			password:    "test12pass",
			maxSequence: 3,
			want:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ContainsSequentialChars(tt.password, tt.maxSequence); got != tt.want {
				t.Errorf("ContainsSequentialChars() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidatePasswordWithCommonChecks(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "Strong password",
			password: "MyS3cur3P@ss",
			wantErr:  false,
		},
		{
			name:     "Common weak password",
			password: "password",
			wantErr:  true,
		},
		{
			name:     "Sequential characters",
			password: "Test123456",
			wantErr:  true,
		},
		{
			name:     "Too short",
			password: "Test@12",
			wantErr:  true,
		},
		{
			name:     "No uppercase",
			password: "test@1234",
			wantErr:  true,
		},
		{
			name:     "Valid complex password",
			password: "C0mpl3xP@ssw0rd",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePasswordWithCommonChecks(tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePasswordWithCommonChecks() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidatePasswordWithPolicy(t *testing.T) {
	// Custom policy requiring special characters
	strictPolicy := PasswordPolicy{
		MinLength:      10,
		MaxLength:      64,
		RequireUpper:   true,
		RequireLower:   true,
		RequireDigit:   true,
		RequireSpecial: true,
	}

	tests := []struct {
		name     string
		password string
		policy   PasswordPolicy
		wantErr  bool
	}{
		{
			name:     "Valid with strict policy",
			password: "MyP@ssw0rd123",
			policy:   strictPolicy,
			wantErr:  false,
		},
		{
			name:     "No special char with strict policy",
			password: "MyPassword123",
			policy:   strictPolicy,
			wantErr:  true,
		},
		{
			name:     "Too short for strict policy",
			password: "Test@123",
			policy:   strictPolicy,
			wantErr:  true,
		},
		{
			name:     "Valid with default policy",
			password: "TestPass123",
			policy:   DefaultPasswordPolicy,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePasswordWithPolicy(tt.password, tt.policy)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePasswordWithPolicy() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Benchmark tests
func BenchmarkValidatePasswordStrength(b *testing.B) {
	password := "Test@1234"
	for i := 0; i < b.N; i++ {
		_ = ValidatePasswordStrength(password)
	}
}

func BenchmarkValidatePasswordWithCommonChecks(b *testing.B) {
	password := "MyS3cur3P@ss"
	for i := 0; i < b.N; i++ {
		_ = ValidatePasswordWithCommonChecks(password)
	}
}

func BenchmarkContainsCommonPassword(b *testing.B) {
	password := "myuniquepassword"
	for i := 0; i < b.N; i++ {
		_ = ContainsCommonPassword(password)
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
