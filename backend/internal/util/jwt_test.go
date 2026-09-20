package util

import "testing"

func TestGenerateAndParseToken(t *testing.T) {
	secret := "test-secret-123456"
	tests := []struct {
		name string
		role string
	}{
		{name: "admin", role: "admin"},
		{name: "member", role: "member"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := GenerateToken(secret, 24, 42, "13800000001", tt.role)
			if err != nil {
				t.Fatalf("GenerateToken() error = %v", err)
			}
			claims, err := ParseToken(secret, token)
			if err != nil {
				t.Fatalf("ParseToken() error = %v", err)
			}
			if claims.UserID != 42 || claims.Role != tt.role {
				t.Fatalf("claims mismatch: %+v", claims)
			}
		})
	}
}

func TestParseToken_WrongSecret(t *testing.T) {
	token, err := GenerateToken("secret-a", 24, 1, "13800000001", "member")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ParseToken("secret-b", token); err == nil {
		t.Fatal("expected error for wrong secret")
	}
}
