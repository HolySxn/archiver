package utils

import "regexp"

func IsValidEmail(email string) bool {
	re := regexp.MustCompile(`^([a-zA-Z0-9]([a-zA-Z0-9_\-\.])*[a-zA-Z0-9]?)@([a-zA-Z0-9_\-\.]+).([a-zA-Z_\.\-]{2,5})$`)
	return re.MatchString(email)
}
