package pkg

import (
	"regexp"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

func IsValidEmail(email string) bool {
	return emailRegex.MatchString(email)
}

func IsValidPassword(password string) bool {
	return len(password) >= 8
}

func IsValidUsername(name string) bool {
	return len(name) >= 3 && len(name) <= 100
}

func IsValidTableNumber(number int) bool {
	return number > 0
}
