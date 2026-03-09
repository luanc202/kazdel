package internal

import (
	"math/rand"
)

func GenerateSlug(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	var slug string
	for i := 0; i < length; i++ {
		slug += string(charset[rand.Intn(len(charset))])
	}
	return slug
}
