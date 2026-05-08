package utils

import (
	"regexp"
	"strings"
)

func ValidateNPWP(npwp string) (bool, string) {
	npwp = strings.TrimSpace(npwp)

	if npwp == "-" {
		return true, ""
	}

	cleanNPWP := strings.ReplaceAll(npwp, ".", "")
	cleanNPWP = strings.ReplaceAll(cleanNPWP, "-", "")
	cleanNPWP = strings.ReplaceAll(cleanNPWP, " ", "")

	isNumeric := regexp.MustCompile(`^[0-9]+$`).MatchString(cleanNPWP)
	if !isNumeric {
		return false, "NPWP must be in the form of numbers"
	}

	length := len(cleanNPWP)

	if length != 15 && length != 16 {
		return false, "The length of the NPWP must be 15 or 16 digits"
	}

	return true, ""
}
