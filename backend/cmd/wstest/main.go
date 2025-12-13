package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
)

const jwtSecret = "adab22111811be224304bd27f82fa85b36424b9a4f2e0be16f4033a7e4e2b646"
const testUserID = "a7e5e31c-ade9-46cb-8b11-083460ae313c"

func main() {
	// 1. Generate a valid JWT token
	token, err := generateToken(testUserID)
	if err != nil {
		log.Fatal("Token generation failed:", err)
	}
	fmt.Println("✓ Token generated:", token[:30]+"...")

	// 2. Connect WebSocket
	ws, err := connectWS(token)
	if err != nil {
		log.Fatal("WebSocket connect failed:", err)
	}
	defer ws.Close()
	fmt.Println("✓ WebSocket connected")

	// 3. Start reading messages in background
	received := make(chan map[string]interface{}, 10)
	go func() {
		for {
			_, data, err := ws.ReadMessage()
			if err != nil {
				fmt.Println("Read error:", err)
				return
			}
			var msg map[string]interface{}
			json.Unmarshal(data, &msg)
			fmt.Printf("✓ RECEIVED: %s\n", string(data))
			received <- msg
		}
	}()

	// 4. Send a test message
	testMsg := map[string]interface{}{
		"t": "msg",
		"p": map[string]interface{}{
			"to":                "fcf454d3-d34a-4765-bc75-e9c3aa4bd9c3",
			"encrypted_content": `{"nonce":"test","ciphertext":"hello from go test"}`,
			"temp_id":           fmt.Sprintf("go-test-%d", time.Now().Unix()),
		},
	}
	data, _ := json.Marshal(testMsg)
	fmt.Printf("→ SENDING: %s\n", string(data))
	if err := ws.WriteMessage(websocket.TextMessage, data); err != nil {
		log.Fatal("Send failed:", err)
	}
	fmt.Println("✓ Message sent")

	// 5. Wait for delivered confirmation
	fmt.Println("Waiting for delivered confirmation...")
	select {
	case msg := <-received:
		if msg["t"] == "delivered" {
			fmt.Println("✓✓✓ DELIVERED confirmation received!")
			p := msg["p"].(map[string]interface{})
			fmt.Printf("  temp_id: %v\n", p["temp_id"])
			if m, ok := p["message"].(map[string]interface{}); ok {
				fmt.Printf("  message.id: %v\n", m["id"])
			}
		} else {
			fmt.Printf("Got different message type: %v\n", msg["t"])
		}
	case <-time.After(5 * time.Second):
		fmt.Println("✗ TIMEOUT - no delivered message received!")
	}
}

func generateToken(userID string) (string, error) {
	claims := jwt.MapClaims{
		"uid": userID, // Must be "uid" not "sub" - matches backend Claims struct
		"exp": time.Now().Add(time.Hour).Unix(),
		"iat": time.Now().Unix(),
		"nbf": time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecret))
}

func connectWS(token string) (*websocket.Conn, error) {
	dialer := websocket.Dialer{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	url := fmt.Sprintf("wss://127.0.0.1:9443/ws?token=%s", token)
	conn, _, err := dialer.Dial(url, nil)
	return conn, err
}
