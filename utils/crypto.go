package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword hashes a password using bcrypt
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(bytes), nil
}

// CheckPasswordHash compares a password with its hash
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GenerateRandomString generates a random string of specified length
func GenerateRandomString(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)

	for i := range result {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", fmt.Errorf("failed to generate random string: %w", err)
		}
		result[i] = charset[num.Int64()]
	}

	return string(result), nil
}

// GenerateRandomNumber generates a random number string of specified length
func GenerateRandomNumber(length int) (string, error) {
	const charset = "0123456789"
	result := make([]byte, length)

	for i := range result {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", fmt.Errorf("failed to generate random number: %w", err)
		}
		result[i] = charset[num.Int64()]
	}

	return string(result), nil
}

// GenerateSecureToken generates a secure random token
func GenerateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate secure token: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// GenerateID generates a unique ID (like ride share code)
func GenerateID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return strings.ToUpper(hex.EncodeToString(bytes))
}

// GenerateShareCode generates a 6-character alphanumeric share code
func GenerateShareCode() string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, 6)

	for i := range result {
		num, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		result[i] = charset[num.Int64()]
	}

	return string(result)
}

// HashSHA256 creates SHA256 hash of input string
func HashSHA256(input string) string {
	hash := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hash[:])
}

// GenerateAPIKey generates a secure API key
func GenerateAPIKey() string {
	token, _ := GenerateSecureToken(32)
	return "idr_" + token
}

// MaskSensitiveData masks sensitive information (like phone numbers, emails)
func MaskSensitiveData(data string, maskChar rune, visibleChars int) string {
	if len(data) <= visibleChars*2 {
		return data
	}

	visible := visibleChars
	masked := strings.Repeat(string(maskChar), len(data)-visible*2)
	return data[:visible] + masked + data[len(data)-visible:]
}

// MaskPhoneNumber masks phone number showing only last 4 digits
func MaskPhoneNumber(phone string) string {
	if len(phone) <= 4 {
		return phone
	}
	return strings.Repeat("*", len(phone)-4) + phone[len(phone)-4:]
}

// MaskEmail masks email showing only first character and domain
func MaskEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return email
	}

	username := parts[0]
	domain := parts[1]

	if len(username) <= 1 {
		return email
	}

	maskedUsername := string(username[0]) + strings.Repeat("*", len(username)-1)
	return maskedUsername + "@" + domain
}
