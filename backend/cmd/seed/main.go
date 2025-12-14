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

	// Service user (小安) config - always created
	serviceNickname := getEnv("SEED_SERVICE_NICKNAME", "小安")
	servicePassword := getEnv("SEED_SERVICE_PASSWORD", "")
	servicePrimaryToken := getEnv("SEED_SERVICE_PRIMARY_TOKEN", "")
	serviceBackupToken := getEnv("SEED_SERVICE_BACKUP_TOKEN", "")
	if servicePassword == "" {
		log.Fatal("SEED_SERVICE_PASSWORD is required (小安's password)")
	}
	if servicePrimaryToken == "" || serviceBackupToken == "" {
		log.Fatal("SEED_SERVICE_PRIMARY_TOKEN and SEED_SERVICE_BACKUP_TOKEN are required.\n" +
			"Generate from admin panel and burn to NFC cards (or use URL directly for testing)")
	}

	// Demo user with NFC cards config - optional
	demoNickname := getEnv("SEED_DEMO_NICKNAME", "")
	demoPassword := getEnv("SEED_DEMO_PASSWORD", "")
	demoPrimaryToken := getEnv("SEED_DEMO_PRIMARY_TOKEN", "")
	demoBackupToken := getEnv("SEED_DEMO_BACKUP_TOKEN", "")

	createDemoUser := demoNickname != "" && demoPassword != "" && demoPrimaryToken != "" && demoBackupToken != ""

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

	// Create service user (小安)
	log.Printf("Creating service user: %s", serviceNickname)
	serviceHash, err := hashPassword(servicePassword)
	if err != nil {
		log.Fatalf("Failed to hash service password: %v", err)
	}

	var serviceUserID string
	err = tx.QueryRow(ctx, `
		INSERT INTO users (nickname, password_hash, public_key)
		VALUES ($1, $2, $3)
		RETURNING id
	`, serviceNickname, serviceHash, "placeholder-will-be-derived-on-first-login").Scan(&serviceUserID)
	if err != nil {
		log.Fatalf("Failed to create service user: %v", err)
	}
	log.Printf("Created service user: %s (ID: %s)", serviceNickname, serviceUserID)

	// Create card pair for service user
	_, err = tx.Exec(ctx, `
		INSERT INTO card_pairs (primary_token, backup_token, expires_at)
		VALUES ($1, $2, NOW() + INTERVAL '365 days')
	`, servicePrimaryToken, serviceBackupToken)
	if err != nil {
		log.Fatalf("Failed to create service card pair: %v", err)
	}

	// Create primary card for service user
	_, err = tx.Exec(ctx, `
		INSERT INTO cards (user_id, card_token, card_type, status, activated_at)
		VALUES ($1, $2, 'primary', 'active', NOW())
	`, serviceUserID, servicePrimaryToken)
	if err != nil {
		log.Fatalf("Failed to create service primary card: %v", err)
	}

	// Create backup card for service user
	_, err = tx.Exec(ctx, `
		INSERT INTO cards (user_id, card_token, card_type, status, activated_at)
		VALUES ($1, $2, 'backup', 'active', NOW())
	`, serviceUserID, serviceBackupToken)
	if err != nil {
		log.Fatalf("Failed to create service backup card: %v", err)
	}

	// Optionally create demo user with NFC cards
	var demoUserID string
	if createDemoUser {
		log.Printf("Creating demo user: %s", demoNickname)
		demoHash, err := hashPassword(demoPassword)
		if err != nil {
			log.Fatalf("Failed to hash demo password: %v", err)
		}

		err = tx.QueryRow(ctx, `
			INSERT INTO users (nickname, password_hash, public_key)
			VALUES ($1, $2, $3)
			RETURNING id
		`, demoNickname, demoHash, "placeholder-will-be-derived-on-first-login").Scan(&demoUserID)
		if err != nil {
			log.Fatalf("Failed to create demo user: %v", err)
		}
		log.Printf("Created demo user: %s (ID: %s)", demoNickname, demoUserID)

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
		`, demoUserID, demoPrimaryToken)
		if err != nil {
			log.Fatalf("Failed to create primary card: %v", err)
		}

		// Create backup card
		_, err = tx.Exec(ctx, `
			INSERT INTO cards (user_id, card_token, card_type, status, activated_at)
			VALUES ($1, $2, 'backup', 'active', NOW())
		`, demoUserID, demoBackupToken)
		if err != nil {
			log.Fatalf("Failed to create backup card: %v", err)
		}

		// Auto-friend demo user with service user
		_, err = tx.Exec(ctx, `
			INSERT INTO friendships (requester_id, addressee_id, status)
			VALUES ($1, $2, 'accepted')
		`, serviceUserID, demoUserID)
		if err != nil {
			log.Fatalf("Failed to create friendship: %v", err)
		}
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		log.Fatalf("Failed to commit: %v", err)
	}

	baseURL := getEnv("BASE_URL", "https://link.mcphub.tw")

	log.Println("========================================")
	log.Println("Seed completed successfully!")
	log.Println("========================================")
	log.Println("")
	log.Println("Service User (小安):")
	log.Printf("  Nickname: %s", serviceNickname)
	log.Printf("  ID:       %s", serviceUserID)
	log.Printf("  Password: %s", servicePassword)
	log.Printf("  Login URL: %s/w/%s", baseURL, servicePrimaryToken)
	log.Println("")
	log.Println("Add to .env:")
	log.Printf("  SERVICE_USER_ID=%s", serviceUserID)
	log.Println("")

	if createDemoUser {
		log.Println("Demo User:")
		log.Printf("  Nickname: %s", demoNickname)
		log.Printf("  ID:       %s", demoUserID)
		log.Printf("  Password: %s", demoPassword)
		log.Printf("  Primary Card: %s", demoPrimaryToken)
		log.Printf("  Backup Card:  %s", demoBackupToken)
		log.Println("")
		log.Println("Now you can:")
		log.Println("1. Tap primary NFC card to login")
		log.Println("2. Enter password to complete login")
	} else {
		log.Println("No demo user created (NFC card tokens not provided)")
		log.Println("To create a demo user, set these environment variables:")
		log.Println("  SEED_DEMO_NICKNAME, SEED_DEMO_PASSWORD")
		log.Println("  SEED_DEMO_PRIMARY_TOKEN, SEED_DEMO_BACKUP_TOKEN")
	}
	log.Println("========================================")
}
