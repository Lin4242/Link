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

	// Demo card pair config - optional (for registration demo flow)
	demoNickname := getEnv("SEED_DEMO_NICKNAME", "Demo")      // Just for display
	demoPassword := getEnv("SEED_DEMO_PASSWORD", "(自行設定)") // Just for display
	demoPrimaryToken := getEnv("SEED_DEMO_PRIMARY_TOKEN", "")
	demoBackupToken := getEnv("SEED_DEMO_BACKUP_TOKEN", "")

	createDemoUser := demoPrimaryToken != "" && demoBackupToken != ""

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

	// Optionally create demo card pair (unregistered - for demo registration flow)
	if createDemoUser {
		log.Printf("Creating demo card pair for: %s (unregistered)", demoNickname)

		// Only create card_pair - no user, no cards
		// User will register by scanning the NFC card
		_, err = tx.Exec(ctx, `
			INSERT INTO card_pairs (primary_token, backup_token, expires_at)
			VALUES ($1, $2, NOW() + INTERVAL '365 days')
		`, demoPrimaryToken, demoBackupToken)
		if err != nil {
			log.Fatalf("Failed to create demo card pair: %v", err)
		}
		log.Println("Demo card pair created (ready for registration)")
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
		log.Println("Demo Card Pair (未註冊):")
		log.Printf("  預設暱稱: %s", demoNickname)
		log.Printf("  預設密碼: %s", demoPassword)
		log.Printf("  主卡 Token: %s", demoPrimaryToken)
		log.Printf("  副卡 Token: %s", demoBackupToken)
		log.Printf("  註冊 URL: %s/w/%s", baseURL, demoPrimaryToken)
		log.Println("")
		log.Println("Demo 流程:")
		log.Println("1. 刷主卡 → 進入註冊頁面")
		log.Println("2. 刷副卡 → 配對成功")
		log.Println("3. 輸入暱稱和密碼完成註冊")
		log.Println("4. 自動加小安為好友")
	} else {
		log.Println("No demo card pair created (tokens not provided)")
		log.Println("To create demo card pair, set:")
		log.Println("  SEED_DEMO_PRIMARY_TOKEN, SEED_DEMO_BACKUP_TOKEN")
	}
	log.Println("========================================")
}
