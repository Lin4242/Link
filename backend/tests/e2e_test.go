package tests

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"link/internal/pkg/cardtoken"
)

const baseURL = "https://localhost:8443"

// Use the same secret as the server default
var cardTokenGen = cardtoken.NewGenerator("default-card-token-secret-change-me")

var client = &http.Client{
	Timeout: 10 * time.Second,
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	},
}

type apiResponse struct {
	Data  json.RawMessage `json:"data,omitempty"`
	Error *struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func doRequest(t *testing.T, method, path string, body interface{}) apiResponse {
	var reqBody io.Reader
	if body != nil {
		jsonBytes, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request: %v", err)
		}
		reqBody = bytes.NewReader(jsonBytes)
	}

	req, err := http.NewRequest(method, baseURL+path, reqBody)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	defer resp.Body.Close()

	var result apiResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		// Some endpoints return plain text
		return apiResponse{}
	}

	return result
}

func doRequestWithToken(t *testing.T, method, path string, body interface{}, token string) apiResponse {
	var reqBody io.Reader
	if body != nil {
		jsonBytes, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request: %v", err)
		}
		reqBody = bytes.NewReader(jsonBytes)
	}

	req, err := http.NewRequest(method, baseURL+path, reqBody)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	defer resp.Body.Close()

	var result apiResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return apiResponse{}
	}

	return result
}

// TestHealthEndpoint 測試健康檢查端點
func TestHealthEndpoint(t *testing.T) {
	resp, err := client.Get(baseURL + "/health")
	if err != nil {
		t.Fatalf("health check failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if string(body) != "OK" {
		t.Errorf("expected 'OK', got '%s'", string(body))
	}
}

// TestCheckCard_InvalidToken 測試檢查無效格式的卡片
func TestCheckCard_InvalidToken(t *testing.T) {
	result := doRequest(t, "GET", "/api/v1/auth/check-card/nonexistent-token-12345", nil)

	if result.Error != nil {
		t.Fatalf("unexpected error: %v", result.Error.Message)
	}

	var data struct {
		Status string `json:"status"`
	}
	if err := json.Unmarshal(result.Data, &data); err != nil {
		t.Fatalf("unmarshal data: %v", err)
	}

	if data.Status != "invalid_token" {
		t.Errorf("expected status 'invalid_token', got '%s'", data.Status)
	}
}

// TestCheckCard_ValidToken 測試檢查有效格式的卡片
func TestCheckCard_ValidToken(t *testing.T) {
	primary, _, err := cardTokenGen.GeneratePair()
	if err != nil {
		t.Fatalf("generate pair: %v", err)
	}

	result := doRequest(t, "GET", "/api/v1/auth/check-card/"+primary, nil)

	if result.Error != nil {
		t.Fatalf("unexpected error: %v", result.Error.Message)
	}

	var data struct {
		Status string `json:"status"`
	}
	if err := json.Unmarshal(result.Data, &data); err != nil {
		t.Fatalf("unmarshal data: %v", err)
	}

	if data.Status != "can_register" {
		t.Errorf("expected status 'can_register', got '%s'", data.Status)
	}
}

// TestDualCardRegistrationFlow 測試雙卡註冊流程
func TestDualCardRegistrationFlow(t *testing.T) {
	// 等待避免觸發 rate limiting
	time.Sleep(2 * time.Second)

	// 使用 cardtoken generator 產生配對的 tokens
	primaryToken, backupToken, err := cardTokenGen.GeneratePair()
	if err != nil {
		t.Fatalf("generate pair: %v", err)
	}

	t.Logf("Primary token: %s", primaryToken)
	t.Logf("Backup token: %s", backupToken)

	// 步驟 1: 檢查主卡可以註冊
	t.Run("CheckPrimaryCanRegister", func(t *testing.T) {
		result := doRequest(t, "GET", "/api/v1/auth/check-card/"+primaryToken, nil)

		if result.Error != nil {
			t.Fatalf("check card failed: %s", result.Error.Message)
		}

		var data struct {
			Status string `json:"status"`
		}
		if err := json.Unmarshal(result.Data, &data); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}

		if data.Status != "can_register" {
			t.Errorf("expected status 'can_register', got '%s'", data.Status)
		}
	})

	// 步驟 2: 直接註冊（不需要配對步驟）
	var token string
	var userID string
	t.Run("Register", func(t *testing.T) {
		result := doRequest(t, "POST", "/api/v1/auth/register", map[string]string{
			"primary_token": primaryToken,
			"backup_token":  backupToken,
			"password":      "TestPassword123!",
			"nickname":      "測試用戶",
			"public_key":    "abcd1234567890abcd1234567890abcd1234567890abcd1234567890abcd1234",
		})

		if result.Error != nil {
			t.Fatalf("register failed: %s", result.Error.Message)
		}

		var data struct {
			User struct {
				ID       string `json:"id"`
				Nickname string `json:"nickname"`
			} `json:"user"`
			Token string `json:"token"`
		}
		if err := json.Unmarshal(result.Data, &data); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}

		if data.Token == "" {
			t.Error("expected token in response")
		}
		if data.User.ID == "" {
			t.Error("expected user id in response")
		}
		if data.User.Nickname != "測試用戶" {
			t.Errorf("expected nickname '測試用戶', got '%s'", data.User.Nickname)
		}

		token = data.Token
		userID = data.User.ID
	})

	// 步驟 3: 使用主卡登入
	t.Run("LoginWithPrimaryCard", func(t *testing.T) {
		result := doRequest(t, "POST", "/api/v1/auth/login", map[string]string{
			"card_token": primaryToken,
			"password":   "TestPassword123!",
		})

		if result.Error != nil {
			t.Fatalf("login failed: %s", result.Error.Message)
		}

		var data struct {
			Token string `json:"token"`
			User  struct {
				ID string `json:"id"`
			} `json:"user"`
		}
		if err := json.Unmarshal(result.Data, &data); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}

		if data.Token == "" {
			t.Error("expected token in response")
		}
		if data.User.ID != userID {
			t.Errorf("expected user id '%s', got '%s'", userID, data.User.ID)
		}
	})

	// 步驟 4: 驗證已登入狀態
	t.Run("GetMe", func(t *testing.T) {
		result := doRequestWithToken(t, "GET", "/api/v1/users/me", nil, token)

		if result.Error != nil {
			t.Fatalf("get me failed: %s", result.Error.Message)
		}

		var data struct {
			Nickname string `json:"nickname"`
		}
		if err := json.Unmarshal(result.Data, &data); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}

		if data.Nickname != "測試用戶" {
			t.Errorf("expected nickname '測試用戶', got '%s'", data.Nickname)
		}
	})

	// 步驟 5: 檢查主卡狀態
	t.Run("CheckPrimaryCard", func(t *testing.T) {
		result := doRequest(t, "GET", "/api/v1/auth/check-card/"+primaryToken, nil)

		var data struct {
			Status string `json:"status"`
		}
		if err := json.Unmarshal(result.Data, &data); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}

		if data.Status != "primary" {
			t.Errorf("expected status 'primary', got '%s'", data.Status)
		}
	})

	// 步驟 6: 檢查附卡狀態
	t.Run("CheckBackupCard", func(t *testing.T) {
		result := doRequest(t, "GET", "/api/v1/auth/check-card/"+backupToken, nil)

		var data struct {
			Status string `json:"status"`
		}
		if err := json.Unmarshal(result.Data, &data); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}

		if data.Status != "backup" {
			t.Errorf("expected status 'backup', got '%s'", data.Status)
		}
	})
}

// TestRegisterWithMismatchedTokens 測試使用不配對的 tokens 註冊
func TestRegisterWithMismatchedTokens(t *testing.T) {
	time.Sleep(2 * time.Second)

	// 產生兩對不同的 tokens
	primary1, _, _ := cardTokenGen.GeneratePair()
	_, backup2, _ := cardTokenGen.GeneratePair()

	result := doRequest(t, "POST", "/api/v1/auth/register", map[string]string{
		"primary_token": primary1,
		"backup_token":  backup2, // 不配對的 backup
		"password":      "TestPassword123!",
		"nickname":      "測試",
		"public_key":    "abcd1234567890abcd1234567890abcd1234567890abcd1234567890abcd1234",
	})

	if result.Error == nil {
		t.Error("expected error for mismatched tokens")
	}
}

// TestBackupCardRevocation 測試附卡登入撤銷主卡
func TestBackupCardRevocation(t *testing.T) {
	time.Sleep(2 * time.Second)

	primaryToken, backupToken, err := cardTokenGen.GeneratePair()
	if err != nil {
		t.Fatalf("generate pair: %v", err)
	}
	password := "RevokeTestPassword123!"

	// 先完成註冊
	registerResult := doRequest(t, "POST", "/api/v1/auth/register", map[string]string{
		"primary_token": primaryToken,
		"backup_token":  backupToken,
		"password":      password,
		"nickname":      "撤銷測試",
		"public_key":    "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
	})
	if registerResult.Error != nil {
		t.Fatalf("register failed: %s", registerResult.Error.Message)
	}

	// 步驟 1: 使用附卡登入 (不確認，應該失敗)
	t.Run("BackupLoginWithoutConfirm", func(t *testing.T) {
		result := doRequest(t, "POST", "/api/v1/auth/login/backup", map[string]interface{}{
			"card_token": backupToken,
			"password":   password,
			"confirm":    false,
		})

		if result.Error == nil {
			t.Error("expected error for backup login without confirm")
		}
	})

	// 步驟 2: 使用附卡登入 (確認撤銷)
	t.Run("BackupLoginWithConfirm", func(t *testing.T) {
		result := doRequest(t, "POST", "/api/v1/auth/login/backup", map[string]interface{}{
			"card_token": backupToken,
			"password":   password,
			"confirm":    true,
		})

		if result.Error != nil {
			t.Fatalf("backup login failed: %s", result.Error.Message)
		}

		var data struct {
			Token string `json:"token"`
		}
		if err := json.Unmarshal(result.Data, &data); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}

		if data.Token == "" {
			t.Error("expected token in response")
		}
	})

	// 步驟 3: 驗證主卡已被撤銷
	t.Run("PrimaryCardRevoked", func(t *testing.T) {
		result := doRequest(t, "GET", "/api/v1/auth/check-card/"+primaryToken, nil)

		var data struct {
			Status string `json:"status"`
		}
		if err := json.Unmarshal(result.Data, &data); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}

		if data.Status != "revoked" {
			t.Errorf("expected status 'revoked', got '%s'", data.Status)
		}
	})

	// 步驟 4: 嘗試使用已撤銷的主卡登入
	t.Run("LoginWithRevokedCard", func(t *testing.T) {
		result := doRequest(t, "POST", "/api/v1/auth/login", map[string]string{
			"card_token": primaryToken,
			"password":   password,
		})

		if result.Error == nil {
			t.Error("expected error for revoked card login")
		}
	})
}

// TestInvalidLogin 測試無效的登入嘗試
func TestInvalidLogin(t *testing.T) {
	t.Run("NonexistentCard", func(t *testing.T) {
		// 用有效格式但未註冊的 token
		primary, _, _ := cardTokenGen.GeneratePair()
		result := doRequest(t, "POST", "/api/v1/auth/login", map[string]string{
			"card_token": primary,
			"password":   "password",
		})

		if result.Error == nil {
			t.Error("expected error for nonexistent card")
		}
	})

	t.Run("WrongPassword", func(t *testing.T) {
		time.Sleep(2 * time.Second)

		primaryToken, backupToken, _ := cardTokenGen.GeneratePair()

		regRes := doRequest(t, "POST", "/api/v1/auth/register", map[string]string{
			"primary_token": primaryToken,
			"backup_token":  backupToken,
			"password":      "CorrectPassword123!",
			"nickname":      "密碼測試",
			"public_key":    "fedcba0987654321fedcba0987654321fedcba0987654321fedcba0987654321",
		})
		if regRes.Error != nil {
			t.Fatalf("register failed: %s", regRes.Error.Message)
		}

		// 嘗試用錯誤密碼登入
		result := doRequest(t, "POST", "/api/v1/auth/login", map[string]string{
			"card_token": primaryToken,
			"password":   "WrongPassword456!",
		})

		if result.Error == nil {
			t.Error("expected error for wrong password")
		}
	})
}

// TestProtectedEndpointsRequireAuth 測試受保護端點需要認證
func TestProtectedEndpointsRequireAuth(t *testing.T) {
	endpoints := []struct {
		method string
		path   string
	}{
		{"GET", "/api/v1/users/me"},
		{"GET", "/api/v1/users/me/cards"},
		{"GET", "/api/v1/friends"},
		{"GET", "/api/v1/conversations"},
	}

	for _, ep := range endpoints {
		t.Run(ep.method+" "+ep.path, func(t *testing.T) {
			req, _ := http.NewRequest(ep.method, baseURL+ep.path, nil)
			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusUnauthorized {
				t.Errorf("expected status 401, got %d", resp.StatusCode)
			}
		})
	}
}
