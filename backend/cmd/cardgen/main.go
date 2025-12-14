package main

import (
	"flag"
	"fmt"
	"os"

	"link/internal/pkg/cardtoken"
)

func main() {
	baseURL := flag.String("url", "https://localhost:5173", "Base URL for the frontend")
	secret := flag.String("secret", "default-card-token-secret-change-me", "Secret key for HMAC (use same as server)")
	count := flag.Int("n", 1, "Number of pairs to generate")
	flag.Parse()

	gen := cardtoken.NewGenerator(*secret)

	fmt.Println("=== LINK 卡片產生器 ===")
	fmt.Printf("Base URL: %s\n", *baseURL)
	fmt.Println()

	for i := 0; i < *count; i++ {
		primary, backup, err := gen.GeneratePair()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error generating pair: %v\n", err)
			os.Exit(1)
		}

		if *count > 1 {
			fmt.Printf("--- Pair %d ---\n", i+1)
		}

		fmt.Println("主卡 (Primary):")
		fmt.Printf("  Token: %s\n", primary)
		fmt.Printf("  URL:   %s/register/start?token=%s\n", *baseURL, primary)
		fmt.Println()

		fmt.Println("副卡 (Backup):")
		fmt.Printf("  Token: %s\n", backup)
		fmt.Printf("  URL:   %s/register/start?token=%s\n", *baseURL, backup)
		fmt.Println()

		if *count > 1 {
			fmt.Println()
		}
	}

	fmt.Println("將以上 URL 寫入 NFC 卡片的 NDEF 記錄即可使用")
}
