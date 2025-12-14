package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
)

const (
	jwtSecret = "adab22111811be224304bd27f82fa85b36424b9a4f2e0be16f4033a7e4e2b646"
	xiaoAnID  = "fcf454d3-d34a-4765-bc75-e9c3aa4bd9c3"
	// Development testing URL - update for your environment
	wsURL     = "wss://localhost:8443/ws" // or use environment variable: os.Getenv("WS_URL")
)

func main() {
	// Generate JWT for 小安
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"uid": xiaoAnID,
		"exp": time.Now().Add(time.Hour).Unix(),
		"iat": time.Now().Unix(),
		"nbf": time.Now().Unix(),
	})
	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		log.Fatal("Failed to generate token:", err)
	}

	fmt.Println("=== 小安上線了 ===")
	fmt.Println("User ID:", xiaoAnID)
	fmt.Println("Connecting to WebSocket...")

	// Connect to WebSocket
	u, _ := url.Parse(wsURL)
	q := u.Query()
	q.Set("token", tokenString)
	u.RawQuery = q.Encode()

	dialer := websocket.Dialer{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	conn, _, err := dialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("WebSocket connection failed:", err)
	}
	defer conn.Close()

	fmt.Println("Connected! Waiting for messages...")
	fmt.Println("Press Ctrl+C to exit")
	fmt.Println("---")

	// Handle Ctrl+C
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	// Read messages
	go func() {
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Println("Read error:", err)
				return
			}

			var msg map[string]interface{}
			if err := json.Unmarshal(message, &msg); err != nil {
				fmt.Println("Raw message:", string(message))
				continue
			}

			msgType := msg["t"].(string)
			timestamp := time.Now().Format("15:04:05")

			payload, _ := msg["p"].(map[string]interface{})

			switch msgType {
			case "msg":
				senderID, _ := payload["sender_id"].(string)
				fmt.Printf("[%s] 📩 收到訊息！\n", timestamp)
				fmt.Printf("  From: %s\n", senderID)
				fmt.Printf("  Content: (encrypted)\n")
				fmt.Println("---")
			case "typing":
				userID, _ := payload["user_id"].(string)
				if len(userID) > 8 {
					userID = userID[:8]
				}
				fmt.Printf("[%s] ✏️ %s 正在輸入...\n", timestamp, userID)
			case "online":
				userID, _ := payload["user_id"].(string)
				if len(userID) > 8 {
					userID = userID[:8]
				}
				fmt.Printf("[%s] 🟢 %s 上線了\n", timestamp, userID)
			case "offline":
				userID, _ := payload["user_id"].(string)
				if len(userID) > 8 {
					userID = userID[:8]
				}
				fmt.Printf("[%s] ⚪ %s 離線了\n", timestamp, userID)
			default:
				fmt.Printf("[%s] 📨 Message type: %s\n", timestamp, msgType)
				fmt.Printf("  Payload: %v\n", payload)
			}
		}
	}()

	<-interrupt
	fmt.Println("\n小安下線了")
}
