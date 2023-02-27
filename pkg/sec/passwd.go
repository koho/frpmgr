package sec

import (
	"crypto/sha1"
	"encoding/base64"
)

// EncryptPassword returns a Base64-encoded string of the hashed password.
func EncryptPassword(password string) string {
	hashed := sha1.Sum([]byte(password))
	return base64.StdEncoding.EncodeToString(hashed[:])
}
