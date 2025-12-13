package cardtoken

import (
	"strings"
	"testing"
)

func TestGeneratePair(t *testing.T) {
	g := NewGenerator("test-secret-key")

	primary, backup, err := g.GeneratePair()
	if err != nil {
		t.Fatalf("GeneratePair failed: %v", err)
	}

	t.Logf("Primary: %s", primary)
	t.Logf("Backup:  %s", backup)

	// Check format
	if !strings.Contains(primary, "-P-") {
		t.Errorf("Primary token should contain -P-, got: %s", primary)
	}
	if !strings.Contains(backup, "-B-") {
		t.Errorf("Backup token should contain -B-, got: %s", backup)
	}

	// Check they share the same pairID
	pairID1, _, _ := g.ParseToken(primary)
	pairID2, _, _ := g.ParseToken(backup)
	if pairID1 != pairID2 {
		t.Errorf("Tokens should have same pairID: %s vs %s", pairID1, pairID2)
	}
}

func TestParseToken(t *testing.T) {
	g := NewGenerator("test-secret-key")

	primary, backup, _ := g.GeneratePair()

	// Test primary
	pairID, cardType, err := g.ParseToken(primary)
	if err != nil {
		t.Fatalf("ParseToken(primary) failed: %v", err)
	}
	if cardType != TypePrimary {
		t.Errorf("Expected type P, got %s", cardType)
	}
	if pairID == "" {
		t.Error("pairID should not be empty")
	}

	// Test backup
	_, cardType, err = g.ParseToken(backup)
	if err != nil {
		t.Fatalf("ParseToken(backup) failed: %v", err)
	}
	if cardType != TypeBackup {
		t.Errorf("Expected type B, got %s", cardType)
	}
}

func TestInvalidToken(t *testing.T) {
	g := NewGenerator("test-secret-key")

	tests := []struct {
		name  string
		token string
	}{
		{"empty", ""},
		{"no dashes", "abcd1234"},
		{"wrong format", "abc-def"},
		{"invalid type", "12345678-X-abcdefgh"},
		{"wrong hmac", "12345678-P-wrongsig"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := g.ParseToken(tt.token)
			if err == nil {
				t.Errorf("Expected error for token: %s", tt.token)
			}
		})
	}
}

func TestGetPairedToken(t *testing.T) {
	g := NewGenerator("test-secret-key")

	primary, backup, _ := g.GeneratePair()

	// From primary, get backup
	pairedBackup, err := g.GetPairedToken(primary)
	if err != nil {
		t.Fatalf("GetPairedToken(primary) failed: %v", err)
	}
	if pairedBackup != backup {
		t.Errorf("Expected %s, got %s", backup, pairedBackup)
	}

	// From backup, get primary
	pairedPrimary, err := g.GetPairedToken(backup)
	if err != nil {
		t.Fatalf("GetPairedToken(backup) failed: %v", err)
	}
	if pairedPrimary != primary {
		t.Errorf("Expected %s, got %s", primary, pairedPrimary)
	}
}

func TestArePaired(t *testing.T) {
	g := NewGenerator("test-secret-key")

	primary1, backup1, _ := g.GeneratePair()
	primary2, backup2, _ := g.GeneratePair()

	// Same pair
	if !g.ArePaired(primary1, backup1) {
		t.Error("primary1 and backup1 should be paired")
	}
	if !g.ArePaired(backup1, primary1) {
		t.Error("backup1 and primary1 should be paired (reverse)")
	}

	// Different pairs
	if g.ArePaired(primary1, backup2) {
		t.Error("primary1 and backup2 should NOT be paired")
	}
	if g.ArePaired(primary1, primary2) {
		t.Error("primary1 and primary2 should NOT be paired")
	}
}

func TestIsPrimaryIsBackup(t *testing.T) {
	g := NewGenerator("test-secret-key")

	primary, backup, _ := g.GeneratePair()

	if !g.IsPrimary(primary) {
		t.Error("primary should be identified as primary")
	}
	if g.IsPrimary(backup) {
		t.Error("backup should NOT be identified as primary")
	}
	if !g.IsBackup(backup) {
		t.Error("backup should be identified as backup")
	}
	if g.IsBackup(primary) {
		t.Error("primary should NOT be identified as backup")
	}
}
