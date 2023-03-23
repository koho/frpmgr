package sec

import "testing"

func TestEncryptPassword(t *testing.T) {
	output := EncryptPassword("123456")
	expected := "fEqNCco3Yq9h5ZUglD3CZJT4lBs="
	if output != expected {
		t.Errorf("Expected: %v, got: %v", expected, output)
	}
}
