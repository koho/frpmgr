package sec

import (
	"bytes"
	"testing"
)

func TestEncrypt(t *testing.T) {
	original := []byte("Test data for DPAPI.")
	e, err := Encrypt(original)
	if err != nil {
		t.Errorf("Error encrypting: %s", err.Error())
	}
	d, err := Decrypt(e)
	if err != nil {
		t.Errorf("Error decrypting: %s", err.Error())
	}
	if !bytes.Equal(d, original) {
		t.Error("Decrypted content does not match original")
	}
}
