package cardtoken

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
)

const (
	// Neutral card types - roles are assigned during registration
	TypeFirst  = "1"
	TypeSecond = "2"

	// Legacy types for backward compatibility
	TypePrimary = "P"
	TypeBackup  = "B"
)

var (
	ErrInvalidToken = errors.New("invalid card token format")
	ErrInvalidHMAC  = errors.New("invalid card token signature")
)

type Generator struct {
	secret []byte
}

func NewGenerator(secret string) *Generator {
	return &Generator{secret: []byte(secret)}
}

// GeneratePair creates a matched pair of card tokens with neutral types
// The first/second distinction is just for pairing - roles (primary/backup) are assigned during registration
func (g *Generator) GeneratePair() (firstToken, secondToken string, err error) {
	pairID, err := randomHex(8)
	if err != nil {
		return "", "", err
	}

	firstToken = g.createToken(pairID, TypeFirst)
	secondToken = g.createToken(pairID, TypeSecond)
	return firstToken, secondToken, nil
}

// createToken generates a token: {pairID}-{type}-{hmac[:8]}
func (g *Generator) createToken(pairID, cardType string) string {
	sig := g.computeHMAC(pairID, cardType)
	return pairID + "-" + cardType + "-" + sig[:8]
}

// computeHMAC calculates HMAC-SHA256 for pair validation
func (g *Generator) computeHMAC(pairID, cardType string) string {
	h := hmac.New(sha256.New, g.secret)
	h.Write([]byte(pairID + cardType))
	return hex.EncodeToString(h.Sum(nil))
}

// ParseToken extracts pairID and cardType from a token, validates HMAC
// Accepts both new format (1/2) and legacy format (P/B)
func (g *Generator) ParseToken(token string) (pairID, cardType string, err error) {
	parts := strings.Split(token, "-")
	if len(parts) != 3 {
		return "", "", ErrInvalidToken
	}

	pairID = parts[0]
	cardType = parts[1]
	sig := parts[2]

	// Accept both new (1/2) and legacy (P/B) formats
	validTypes := []string{TypeFirst, TypeSecond, TypePrimary, TypeBackup}
	isValid := false
	for _, t := range validTypes {
		if cardType == t {
			isValid = true
			break
		}
	}
	if !isValid {
		return "", "", ErrInvalidToken
	}

	expectedSig := g.computeHMAC(pairID, cardType)
	if sig != expectedSig[:8] {
		return "", "", ErrInvalidHMAC
	}

	return pairID, cardType, nil
}

// GetPairedToken returns the paired token for a given token
// Works with both new (1/2) and legacy (P/B) formats
func (g *Generator) GetPairedToken(token string) (string, error) {
	pairID, cardType, err := g.ParseToken(token)
	if err != nil {
		return "", err
	}

	switch cardType {
	case TypeFirst:
		return g.createToken(pairID, TypeSecond), nil
	case TypeSecond:
		return g.createToken(pairID, TypeFirst), nil
	case TypePrimary:
		return g.createToken(pairID, TypeBackup), nil
	case TypeBackup:
		return g.createToken(pairID, TypePrimary), nil
	}
	return "", ErrInvalidToken
}

// IsPrimary checks if token is a primary card (legacy P type)
// For new neutral format (1/2), this returns false - role is determined during registration
func (g *Generator) IsPrimary(token string) bool {
	_, cardType, err := g.ParseToken(token)
	return err == nil && cardType == TypePrimary
}

// IsBackup checks if token is a backup card (legacy B type)
// For new neutral format (1/2), this returns false - role is determined during registration
func (g *Generator) IsBackup(token string) bool {
	_, cardType, err := g.ParseToken(token)
	return err == nil && cardType == TypeBackup
}

// IsNeutralFormat checks if token uses the new neutral format (1/2)
func (g *Generator) IsNeutralFormat(token string) bool {
	_, cardType, err := g.ParseToken(token)
	return err == nil && (cardType == TypeFirst || cardType == TypeSecond)
}

// ArePaired checks if two tokens are a valid pair
func (g *Generator) ArePaired(token1, token2 string) bool {
	pairID1, type1, err1 := g.ParseToken(token1)
	pairID2, type2, err2 := g.ParseToken(token2)

	if err1 != nil || err2 != nil {
		return false
	}

	// Must have same pairID and different types
	// For new format: 1 and 2 are paired
	// For legacy format: P and B are paired
	// Cross-format pairing is not allowed
	if pairID1 != pairID2 {
		return false
	}

	// Check same format and different types
	isNewFormat1 := type1 == TypeFirst || type1 == TypeSecond
	isNewFormat2 := type2 == TypeFirst || type2 == TypeSecond

	if isNewFormat1 != isNewFormat2 {
		return false // Can't mix formats
	}

	return type1 != type2
}

func randomHex(n int) (string, error) {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
