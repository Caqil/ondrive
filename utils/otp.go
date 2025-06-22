package utils

import (
	"fmt"
	"strconv"
	"time"
)

// OTPConfig holds configuration for OTP generation and validation
type OTPConfig struct {
	Length      int           `json:"length"`
	ExpiryTime  time.Duration `json:"expiry_time"`
	MaxAttempts int           `json:"max_attempts"`
}

// OTPData holds OTP information
type OTPData struct {
	Code      string    `json:"code"`
	Phone     string    `json:"phone"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
	Attempts  int       `json:"attempts"`
	Used      bool      `json:"used"`
}

// OTPManager manages OTP operations
type OTPManager struct {
	config OTPConfig
	store  map[string]*OTPData // In production, use Redis or database
}

// NewOTPManager creates a new OTP manager
func NewOTPManager(config OTPConfig) *OTPManager {
	if config.Length == 0 {
		config.Length = 6
	}
	if config.ExpiryTime == 0 {
		config.ExpiryTime = 10 * time.Minute
	}
	if config.MaxAttempts == 0 {
		config.MaxAttempts = 3
	}

	return &OTPManager{
		config: config,
		store:  make(map[string]*OTPData),
	}
}

// GenerateOTP generates a new OTP for a phone number
func (o *OTPManager) GenerateOTP(phone string) (*OTPData, error) {
	// Generate random numeric code
	code, err := GenerateRandomNumber(o.config.Length)
	if err != nil {
		return nil, fmt.Errorf("failed to generate OTP: %w", err)
	}

	now := time.Now()
	otpData := &OTPData{
		Code:      code,
		Phone:     phone,
		CreatedAt: now,
		ExpiresAt: now.Add(o.config.ExpiryTime),
		Attempts:  0,
		Used:      false,
	}

	// Store OTP (in production, use Redis with expiration)
	o.store[phone] = otpData

	return otpData, nil
}

// ValidateOTP validates an OTP code for a phone number
func (o *OTPManager) ValidateOTP(phone, code string) error {
	otpData, exists := o.store[phone]
	if !exists {
		return fmt.Errorf("no OTP found for phone number")
	}

	// Check if OTP is already used
	if otpData.Used {
		return fmt.Errorf("OTP has already been used")
	}

	// Check if OTP is expired
	if time.Now().After(otpData.ExpiresAt) {
		delete(o.store, phone) // Clean up expired OTP
		return fmt.Errorf("OTP has expired")
	}

	// Increment attempts
	otpData.Attempts++

	// Check max attempts
	if otpData.Attempts > o.config.MaxAttempts {
		delete(o.store, phone) // Clean up after max attempts
		return fmt.Errorf("maximum OTP attempts exceeded")
	}

	// Validate code
	if otpData.Code != code {
		return fmt.Errorf("invalid OTP code")
	}

	// Mark as used
	otpData.Used = true

	// Clean up used OTP
	delete(o.store, phone)

	return nil
}

// ResendOTP generates a new OTP for the same phone number
func (o *OTPManager) ResendOTP(phone string) (*OTPData, error) {
	// Check if there's an existing OTP
	if existing, exists := o.store[phone]; exists {
		// Check if enough time has passed since last generation (prevent spam)
		if time.Since(existing.CreatedAt) < 1*time.Minute {
			return nil, fmt.Errorf("please wait before requesting a new OTP")
		}
	}

	// Generate new OTP
	return o.GenerateOTP(phone)
}

// GetOTPInfo returns OTP information without revealing the code
func (o *OTPManager) GetOTPInfo(phone string) (*OTPData, error) {
	otpData, exists := o.store[phone]
	if !exists {
		return nil, fmt.Errorf("no OTP found for phone number")
	}

	// Return copy without the actual code
	info := &OTPData{
		Phone:     otpData.Phone,
		CreatedAt: otpData.CreatedAt,
		ExpiresAt: otpData.ExpiresAt,
		Attempts:  otpData.Attempts,
		Used:      otpData.Used,
	}

	return info, nil
}

// CleanupExpiredOTPs removes expired OTPs from store
func (o *OTPManager) CleanupExpiredOTPs() {
	now := time.Now()
	for phone, otpData := range o.store {
		if now.After(otpData.ExpiresAt) {
			delete(o.store, phone)
		}
	}
}

// GenerateNumericOTP generates a numeric OTP of specified length
func GenerateNumericOTP(length int) (string, error) {
	if length <= 0 {
		length = 6
	}

	code := ""
	for i := 0; i < length; i++ {
		digit, err := GenerateRandomNumber(1)
		if err != nil {
			return "", err
		}
		code += digit
	}

	return code, nil
}

// FormatOTPForSMS formats OTP code for SMS sending
func FormatOTPForSMS(code, appName string) string {
	return fmt.Sprintf("Your %s verification code is: %s. Valid for 10 minutes. Do not share this code.", appName, code)
}

// IsValidOTPFormat checks if OTP format is valid
func IsValidOTPFormat(code string, expectedLength int) bool {
	if len(code) != expectedLength {
		return false
	}

	// Check if all characters are digits
	_, err := strconv.Atoi(code)
	return err == nil
}

// MaskOTPForLogging masks OTP for logging purposes
func MaskOTPForLogging(code string) string {
	if len(code) <= 2 {
		return code
	}
	return code[:1] + "***" + code[len(code)-1:]
}
