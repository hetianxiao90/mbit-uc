package util

import (
	"fmt"
	"testing"
)

func TestEncryptionEmail(t *testing.T) {
	email := EncryptionEmail("a@163.com")
	email1 := EncryptionEmail("aaaa@163.com")
	email2 := EncryptionEmail("a111@163.com")
	fmt.Println(email, email1, email2)
}
func TestGenerateSalt(t *testing.T) {
	for i := 0; i < 100; i++ {
		slat, _ := GenerateSalt(12)

		fmt.Println(slat)
	}

}
