package password

import (
	"strings"
	"testing"
)

func TestHash(t *testing.T) {
	password := "TestPassword123!"

	hash, err := Hash(password)
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}

	// 驗證 hash 格式
	if !strings.HasPrefix(hash, "$argon2id$") {
		t.Errorf("Hash() format error, got = %v", hash)
	}

	parts := strings.Split(hash, "$")
	if len(parts) != 6 {
		t.Errorf("Hash() should have 6 parts, got = %d", len(parts))
	}
}

func TestHash_DifferentSalts(t *testing.T) {
	password := "SamePassword123"

	hash1, _ := Hash(password)
	hash2, _ := Hash(password)

	if hash1 == hash2 {
		t.Error("Hash() should generate different hashes for same password (due to random salt)")
	}
}

func TestVerify_ValidPassword(t *testing.T) {
	password := "ValidPassword123!"

	hash, err := Hash(password)
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}

	valid, err := Verify(password, hash)
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}

	if !valid {
		t.Error("Verify() should return true for valid password")
	}
}

func TestVerify_InvalidPassword(t *testing.T) {
	password := "CorrectPassword123!"
	wrongPassword := "WrongPassword123!"

	hash, _ := Hash(password)

	valid, err := Verify(wrongPassword, hash)
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}

	if valid {
		t.Error("Verify() should return false for invalid password")
	}
}

func TestVerify_InvalidHashFormat(t *testing.T) {
	_, err := Verify("password", "invalid-hash-format")

	if err != ErrInvalidHash {
		t.Errorf("Verify() should return ErrInvalidHash for invalid hash, got = %v", err)
	}
}

func TestVerify_EmptyPassword(t *testing.T) {
	hash, _ := Hash("")

	valid, err := Verify("", hash)
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}

	if !valid {
		t.Error("Verify() should return true for matching empty password")
	}
}

func TestVerify_UnicodePassword(t *testing.T) {
	password := "密碼測試123🔐"

	hash, err := Hash(password)
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}

	valid, err := Verify(password, hash)
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}

	if !valid {
		t.Error("Verify() should work with unicode passwords")
	}
}

func TestVerify_LongPassword(t *testing.T) {
	// 測試長密碼
	password := strings.Repeat("A", 1000)

	hash, err := Hash(password)
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}

	valid, err := Verify(password, hash)
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}

	if !valid {
		t.Error("Verify() should work with long passwords")
	}
}

func TestDefaultParams(t *testing.T) {
	p := DefaultParams()

	if p.Memory != 64*1024 {
		t.Errorf("DefaultParams().Memory = %d, want %d", p.Memory, 64*1024)
	}

	if p.Iterations != 3 {
		t.Errorf("DefaultParams().Iterations = %d, want %d", p.Iterations, 3)
	}

	if p.SaltLength != 16 {
		t.Errorf("DefaultParams().SaltLength = %d, want %d", p.SaltLength, 16)
	}

	if p.KeyLength != 32 {
		t.Errorf("DefaultParams().KeyLength = %d, want %d", p.KeyLength, 32)
	}
}

func BenchmarkHash(b *testing.B) {
	password := "BenchmarkPassword123!"

	for i := 0; i < b.N; i++ {
		_, _ = Hash(password)
	}
}

func BenchmarkVerify(b *testing.B) {
	password := "BenchmarkPassword123!"
	hash, _ := Hash(password)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Verify(password, hash)
	}
}
