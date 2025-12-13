package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"os"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/nacl/box"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

const xiaoAnID = "fcf454d3-d34a-4765-bc75-e9c3aa4bd9c3"

func main() {
	_ = godotenv.Load()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://app:secret@localhost:5432/link?sslmode=disable"
	}

	// Generate password hash for "4242"
	password := "4242"
	hash := hashPassword(password)
	fmt.Println("Password hash generated for '4242'")

	// Generate NaCl key pair
	publicKey, privateKey, err := box.GenerateKey(rand.Reader)
	if err != nil {
		log.Fatal("Failed to generate key pair:", err)
	}

	publicKeyB64 := base64.StdEncoding.EncodeToString(publicKey[:])
	privateKeyB64 := base64.StdEncoding.EncodeToString(privateKey[:])

	fmt.Println("\n=== 小安 Setup ===")
	fmt.Println("User ID:", xiaoAnID)
	fmt.Println("Password: 4242")
	fmt.Println("Public Key:", publicKeyB64)
	fmt.Println("Private Key (save this!):", privateKeyB64)

	// Update database
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, dbURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer conn.Close(ctx)

	_, err = conn.Exec(ctx, `
		UPDATE users
		SET password_hash = $1, public_key = $2
		WHERE id = $3
	`, hash, publicKeyB64, xiaoAnID)
	if err != nil {
		log.Fatal("Failed to update user:", err)
	}

	fmt.Println("\n✅ Database updated successfully!")
	fmt.Println("\nTo enable decryption, 小安 needs to import the private key.")
	fmt.Println("Run the following in browser console after login:")
	fmt.Printf("\n// Import private key for 小安\n")
	fmt.Printf("const secretKey = Uint8Array.from(atob('%s').split('').map(c => c.charCodeAt(0)));\n", privateKeyB64)
	fmt.Println("// Then call saveSecretKey with password '4242'")
}

func hashPassword(password string) string {
	salt := make([]byte, 16)
	rand.Read(salt)

	hash := argon2.IDKey([]byte(password), salt, 3, 64*1024, 4, 32)

	return fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		64*1024,
		3,
		4,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	)
}
