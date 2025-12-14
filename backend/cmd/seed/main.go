package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"runtime"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/argon2"
)

// Argon2id params (matching backend/internal/pkg/password/argon2.go)
type Params struct {
	Memory      uint32
	Iterations  uint32
	Parallelism uint8
	SaltLength  uint32
	KeyLength   uint32
}

func defaultParams() *Params {
	return &Params{
		Memory:      64 * 1024, // 64 MB
		Iterations:  3,
		Parallelism: uint8(runtime.NumCPU()),
		SaltLength:  16,
		KeyLength:   32,
	}
}

func hashPassword(password string) (string, error) {
	params := defaultParams()
	salt := make([]byte, params.SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	hash := argon2.IDKey(
		[]byte(password),
		salt,
		params.Iterations,
		params.Memory,
		params.Parallelism,
		params.KeyLength,
	)

	return fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		params.Memory,
		params.Iterations,
		params.Parallelism,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	), nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func main() {
	// Configuration from environment
	dbURL := getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/link?sslmode=disable")

	// Demo user config
	demoNickname := getEnv("SEED_DEMO_NICKNAME", "小安")
	demoPassword := getEnv("SEED_DEMO_PASSWORD", "demo1234")
	demoPrimaryToken := getEnv("SEED_DEMO_PRIMARY_TOKEN", "")
	demoBackupToken := getEnv("SEED_DEMO_BACKUP_TOKEN", "")

	if demoPrimaryToken == "" || demoBackupToken == "" {
		log.Fatal("SEED_DEMO_PRIMARY_TOKEN and SEED_DEMO_BACKUP_TOKEN are required.\n" +
			"These should match the tokens burned to your NFC cards.\n" +
			"Example: SEED_DEMO_PRIMARY_TOKEN=ABC123 SEED_DEMO_BACKUP_TOKEN=DEF456 go run ./cmd/seed")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Connect to database
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	// Test connection
	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("Connected to database")

	// Start transaction
	tx, err := pool.Begin(ctx)
	if err != nil {
		log.Fatalf("Failed to begin transaction: %v", err)
	}
	defer tx.Rollback(ctx)

	// Clear existing data (in correct order due to foreign keys)
	log.Println("Clearing existing data...")
	tables := []string{"messages", "conversations", "sessions", "cards", "card_pairs", "friendships", "users"}
	for _, table := range tables {
		_, err := tx.Exec(ctx, fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table))
		if err != nil {
			log.Fatalf("Failed to truncate %s: %v", table, err)
		}
	}

	// Hash password
	log.Printf("Creating demo user: %s", demoNickname)
	passwordHash, err := hashPassword(demoPassword)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	// Create demo user
	var userID string
	err = tx.QueryRow(ctx, `
		INSERT INTO users (nickname, password_hash, public_key)
		VALUES ($1, $2, $3)
		RETURNING id
	`, demoNickname, passwordHash, "placeholder-will-be-derived-on-first-login").Scan(&userID)
	if err != nil {
		log.Fatalf("Failed to create user: %v", err)
	}
	log.Printf("Created user: %s (ID: %s)", demoNickname, userID)

	// Create card pair
	log.Println("Creating card pair...")
	_, err = tx.Exec(ctx, `
		INSERT INTO card_pairs (primary_token, backup_token, expires_at)
		VALUES ($1, $2, NOW() + INTERVAL '365 days')
	`, demoPrimaryToken, demoBackupToken)
	if err != nil {
		log.Fatalf("Failed to create card pair: %v", err)
	}

	// Create primary card
	_, err = tx.Exec(ctx, `
		INSERT INTO cards (user_id, card_token, card_type, status, activated_at)
		VALUES ($1, $2, 'primary', 'active', NOW())
	`, userID, demoPrimaryToken)
	if err != nil {
		log.Fatalf("Failed to create primary card: %v", err)
	}

	// Create backup card
	_, err = tx.Exec(ctx, `
		INSERT INTO cards (user_id, card_token, card_type, status, activated_at)
		VALUES ($1, $2, 'backup', 'active', NOW())
	`, userID, demoBackupToken)
	if err != nil {
		log.Fatalf("Failed to create backup card: %v", err)
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		log.Fatalf("Failed to commit: %v", err)
	}

	log.Println("========================================")
	log.Println("Seed completed successfully!")
	log.Println("========================================")
	log.Printf("Demo User: %s", demoNickname)
	log.Printf("Password:  %s", demoPassword)
	log.Printf("Primary Card Token: %s", demoPrimaryToken)
	log.Printf("Backup Card Token:  %s", demoBackupToken)
	log.Println("")
	log.Println("Now you can:")
	log.Println("1. Tap primary NFC card to login")
	log.Println("2. Enter password to complete login")
	log.Println("========================================")
}
