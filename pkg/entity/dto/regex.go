package dto

import "regexp"

var (
	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9._]+$`)
	emailRegex    = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
)
