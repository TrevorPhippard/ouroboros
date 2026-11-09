package services

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func generateID() string {
	return uuid.NewString()
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func CheckPassword(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func GenerateToken(userID string, secret []byte) (string, error) {
	if userID == "" || len(secret) == 0 {
		return "", errors.New("user ID and JWT secret are required")
	}

	encode := base64.RawURLEncoding.EncodeToString
	header, err := json.Marshal(struct {
		Algorithm string `json:"alg"`
		Type      string `json:"typ"`
	}{Algorithm: "HS256", Type: "JWT"})
	if err != nil {
		return "", err
	}
	claims, err := json.Marshal(struct {
		Subject   string `json:"sub"`
		ExpiresAt int64  `json:"exp"`
	}{Subject: userID, ExpiresAt: time.Now().Add(24 * time.Hour).Unix()})
	if err != nil {
		return "", err
	}

	unsigned := encode(header) + "." + encode(claims)
	mac := hmac.New(sha256.New, secret)
	_, _ = mac.Write([]byte(unsigned))
	return unsigned + "." + encode(mac.Sum(nil)), nil
}
