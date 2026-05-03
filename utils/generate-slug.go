// helper/slug.go
package utils

import (
	"strings"

	"github.com/google/uuid"
)

func GenerateSlug(title string) string {
	t := strings.ToLower(title)
	t = strings.ReplaceAll(t, " ", "-")
	// Tambahkan sedikit random string di belakang agar slug tidak duplikat
	return t + "-" + uuid.New().String()[0:5]
}
