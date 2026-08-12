package entity

import "time"

type Setting struct {
	Key       string
	Value     string
	UpdatedAt time.Time
}

func NewSetting(key, value string) *Setting {
	return &Setting{
		Key:       key,
		Value:     value,
		UpdatedAt: time.Now(),
	}
}

const (
	SettingSignupEnabled          = "signup_enabled"
	SettingEmailValidationEnabled = "email_validation_enabled"
)
