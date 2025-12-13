package token

import (
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const testSecret = "this-is-a-very-secure-secret-key-for-testing-purposes"

func TestNewManager(t *testing.T) {
	m := NewManager(testSecret, time.Hour)

	if m == nil {
		t.Error("NewManager() should not return nil")
	}
}

func TestNewManager_ShortSecret_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("NewManager() should panic with short secret")
		}
	}()

	NewManager("short", time.Hour)
}

func TestGenerate(t *testing.T) {
	m := NewManager(testSecret, time.Hour)
	userID := "user-123"

	token, err := m.Generate(userID)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if token == "" {
		t.Error("Generate() should return non-empty token")
	}

	// JWT 應該有三個部分
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Errorf("Generate() token should have 3 parts, got = %d", len(parts))
	}
}

func TestVerify_ValidToken(t *testing.T) {
	m := NewManager(testSecret, time.Hour)
	userID := "user-456"

	token, _ := m.Generate(userID)

	claims, err := m.Verify(token)
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("Verify() UserID = %v, want %v", claims.UserID, userID)
	}
}

func TestVerify_ExpiredToken(t *testing.T) {
	// 使用 -1 秒過期時間建立已過期的 token
	m := NewManager(testSecret, -time.Second)
	token, _ := m.Generate("user-789")

	_, err := m.Verify(token)

	if err != ErrExpiredToken {
		t.Errorf("Verify() should return ErrExpiredToken, got = %v", err)
	}
}

func TestVerify_InvalidSignature(t *testing.T) {
	m1 := NewManager(testSecret, time.Hour)
	m2 := NewManager("another-very-secure-secret-key-for-testing", time.Hour)

	token, _ := m1.Generate("user-test")

	_, err := m2.Verify(token)

	if err != ErrInvalidSignature {
		t.Errorf("Verify() should return ErrInvalidSignature, got = %v", err)
	}
}

func TestVerify_InvalidToken(t *testing.T) {
	m := NewManager(testSecret, time.Hour)

	_, err := m.Verify("invalid-token-string")

	if err != ErrInvalidToken {
		t.Errorf("Verify() should return ErrInvalidToken, got = %v", err)
	}
}

func TestVerify_TamperedToken(t *testing.T) {
	m := NewManager(testSecret, time.Hour)
	token, _ := m.Generate("user-123")

	// 篡改 token
	parts := strings.Split(token, ".")
	parts[1] = parts[1] + "x"
	tamperedToken := strings.Join(parts, ".")

	_, err := m.Verify(tamperedToken)

	if err == nil {
		t.Error("Verify() should return error for tampered token")
	}
}

func TestVerify_AlgorithmConfusionAttack(t *testing.T) {
	m := NewManager(testSecret, time.Hour)

	// 嘗試使用不同的演算法建立 token
	claims := &Claims{
		UserID: "attacker",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}

	// 使用 none 演算法
	noneToken := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	tokenStr, _ := noneToken.SignedString(jwt.UnsafeAllowNoneSignatureType)

	_, err := m.Verify(tokenStr)

	if err == nil || err == ErrExpiredToken {
		t.Error("Verify() should reject 'none' algorithm tokens")
	}
}

func TestVerify_EmptyToken(t *testing.T) {
	m := NewManager(testSecret, time.Hour)

	_, err := m.Verify("")

	if err != ErrInvalidToken {
		t.Errorf("Verify() should return ErrInvalidToken for empty token, got = %v", err)
	}
}

func TestGenerate_MultipleTimes(t *testing.T) {
	m := NewManager(testSecret, time.Hour)
	userID := "user-multi"

	tokens := make(map[string]bool)
	for i := 0; i < 100; i++ {
		token, _ := m.Generate(userID)
		if tokens[token] {
			// Token 可能相同因為時間戳一樣，這是可接受的
			continue
		}
		tokens[token] = true
	}

	// 所有 token 都應該能驗證
	for token := range tokens {
		claims, err := m.Verify(token)
		if err != nil {
			t.Errorf("Verify() failed for generated token: %v", err)
		}
		if claims.UserID != userID {
			t.Errorf("Claims.UserID = %v, want %v", claims.UserID, userID)
		}
	}
}

func BenchmarkGenerate(b *testing.B) {
	m := NewManager(testSecret, time.Hour)

	for i := 0; i < b.N; i++ {
		_, _ = m.Generate("benchmark-user")
	}
}

func BenchmarkVerify(b *testing.B) {
	m := NewManager(testSecret, time.Hour)
	token, _ := m.Generate("benchmark-user")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = m.Verify(token)
	}
}
