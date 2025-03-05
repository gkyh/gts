package gts

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"time"
	
)
const (
	ExpiresTime = 24
)
// 错误定义
var (
	ErrTokenExpired   = errors.New("token has expired")
	ErrInvalidToken   = errors.New("invalid token")
	ErrTokenMalformed = errors.New("token is malformed")
)

// 配置结构体，支持更灵活的配置
type TokenConfig struct {
	Issuer         string
	Subject        string
	Audience       string
	ExpirationTime time.Duration
	NotBefore      time.Duration
}

// JWT管理器
type JWTManager struct {
	secretKey []byte
	mu        sync.RWMutex
	// 可以添加令牌黑名单
	blacklist map[string]time.Time
}

// Payload 使用更多标准字段
// omitempty 是 Go 语言中 JSON 编码的一个特殊标签，它的作用是在序列化 JSON 时，
// 如果对应的字段为零值（nil、空字符串、0、false 等），则不会被包含在生成的 JSON 中。
type Payload struct {
	Issuer   string `json:"iss,omitempty"`
	Subject  string `json:"sub,omitempty"`
	Audience string `json:"aud,omitempty"`

	IssuedAt  int64 `json:"iat,omitempty"`
	ExpiresAt int64 `json:"exp,omitempty"`
	NotBefore int64 `json:"nbf,omitempty"`

}

//var jwtManager *JWTManager
//func init() {
//	jwtManager = NewJWTManager("ff03kdjhj3-dk93ijj393ikff33")
//}

// 创建新的JWT管理器
func NewJWTManager(secretKey string) *JWTManager {
	return &JWTManager{
		secretKey: []byte(secretKey),
		blacklist: make(map[string]time.Time),
	}
}

// 生成令牌
func (m *JWTManager) GenerateToken(config TokenConfig, customClaims map[string]interface{}) (string, error) {
	
	now := time.Now()
	// 基础payload
	payload := Payload{
		Issuer:    config.Issuer,
		Subject:   config.Subject,
		Audience:  config.Audience,
		IssuedAt:  now.Unix(),
		ExpiresAt: now.Add(config.ExpirationTime).Unix(),
		NotBefore: now.Add(config.NotBefore).Unix(),
	}

	// 合并自定义claims
	payloadMap := make(map[string]interface{})

	// 通过JSON转换，确保类型安全
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	err = json.Unmarshal(jsonPayload, &payloadMap)
	if err != nil {
		return "", err
	}

	// 添加自定义claims
	for k, v := range customClaims {
		payloadMap[k] = v
	}

	// Header和签名逻辑保持不变
	header := map[string]string{
		"alg": "HS256",
		"typ": "JWT",
	}

	// 编码header
	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", err
	}
	encodedHeader := base64URLEncode(headerJSON)

	// 编码payload
	payloadJSON, err := json.Marshal(payloadMap)
	if err != nil {
		return "", err
	}
	encodedPayload := base64URLEncode(payloadJSON)

	// 创建签名部分
	unsignedToken := encodedHeader + "." + encodedPayload
	signature := hmacSHA256(unsignedToken, m.secretKey)
	encodedSignature := base64URLEncode(signature)

	return unsignedToken + "." + encodedSignature, nil
}

// 验证令牌
func (m *JWTManager) VerifyToken(tokenString string) (map[string]interface{}, error) {
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return nil, ErrTokenMalformed
	}

	// 检查是否在黑名单
	m.mu.RLock()
	_, exists := m.blacklist[tokenString]
	m.mu.RUnlock()
	if exists {
		return nil, ErrInvalidToken
	}

	// 验证签名
	unsignedToken := parts[0] + "." + parts[1]
	signature := hmacSHA256(unsignedToken, m.secretKey)
	if base64URLEncode(signature) != parts[2] {
		return nil, ErrInvalidToken
	}

	// 解码payload
	payloadJSON, err := base64URLDecode(parts[1])
	if err != nil {
		return nil, err
	}

  // 解析所有claims
  var claims map[string]interface{}
  if err := json.Unmarshal(payloadJSON, &claims); err != nil {
      return nil, err
  }

  // 检查过期时间
  now := time.Now().Unix()
  if exp, ok := claims["exp"].(float64); ok && int64(exp) < now {
      return nil, ErrTokenExpired
  }

  // 检查生效时间
  if nbf, ok := claims["nbf"].(float64); ok && int64(nbf) > now {
       return nil, errors.New("token is not yet valid")
  }

  return claims, nil
}


// 吊销令牌
func (m *JWTManager) RevokeToken(tokenString string, duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.blacklist[tokenString] = time.Now().Add(duration)
}

// 清理过期的黑名单令牌
func (m *JWTManager) CleanupBlacklist() {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Now()
	for token, expiresAt := range m.blacklist {
		if expiresAt.Before(now) {
			delete(m.blacklist, token)
		}
	}
}

// 辅助函数：HMAC-SHA256签名
func hmacSHA256(message string, secret []byte) []byte {
	h := hmac.New(sha256.New, secret)
	h.Write([]byte(message))
	return h.Sum(nil)
}

// Base64URL编码（无填充）
func base64URLEncode(data []byte) string {
	encoded := base64.StdEncoding.EncodeToString(data)
	encoded = strings.ReplaceAll(encoded, "+", "-")
	encoded = strings.ReplaceAll(encoded, "/", "_")
	return strings.TrimRight(encoded, "=")
}

// Base64URL解码
func base64URLDecode(encoded string) ([]byte, error) {
	// 添加必要的填充
	padded := encoded
	switch len(padded) % 4 {
	case 2:
		padded += "=="
	case 3:
		padded += "="
	}

	// 替换URL安全字符
	padded = strings.ReplaceAll(padded, "-", "+")
	padded = strings.ReplaceAll(padded, "_", "/")

	return base64.StdEncoding.DecodeString(padded)
}

// 使用示例
/*
// 创建JWT管理器
manager := NewJWTManager("your-secret-key")

// 生成令牌
token, err := manager.GenerateToken(TokenConfig{
    Issuer:         "MyApp",
    Subject:        "user123",
    ExpirationTime: time.Hour * 24,
}, map[string]interface{}{
    "role": "admin"
})

// 验证令牌
payload, err := manager.VerifyToken(token)

// 吊销令牌
manager.RevokeToken(token, time.Hour)

func ExampleUsage() {
	// 创建JWT管理器
	manager := NewJWTManager("your-256-bit-secret")

	// 生成令牌配置
	config := TokenConfig{
		Issuer:         "MyApp",
		Subject:        "user123",
		Audience:       "web-clients",
		ExpirationTime: time.Hour * 24,
		NotBefore:      0,
	}

	// 自定义claims
	customClaims := map[string]interface{}{
		"user_role": "admin",
	}

	// 生成令牌
	token, err := manager.GenerateToken(config, customClaims)
	if err != nil {
		fmt.Println("Token generation error:", err)
		return
	}

	// 验证令牌
	payload, err := manager.VerifyToken(token)
	if err != nil {
		fmt.Println("Token verification error:", err)
		return
	}

	fmt.Printf("Token validated for user: %s\n", payload.Subject)
}
*/
