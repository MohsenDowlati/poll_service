package validation

import "regexp"

var phoneRegex = regexp.MustCompile(`^\+?[0-9]{7,15}$`)

// Phone validates that a phone number matches a simplified E.164 pattern.
func Phone(value string) bool {
	return phoneRegex.MatchString(value)
}
