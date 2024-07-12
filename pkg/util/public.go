package util

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math/big"
	"regexp"
	"sync"
)

var mu sync.Mutex

func RandInt64(min, max int64) int64 {
	mu.Lock()
	defer mu.Unlock()
	if min >= max {
		return min
	}

	maxBigInt := big.NewInt(max - min)
	i, _ := rand.Int(rand.Reader, maxBigInt)
	iInt64 := i.Int64() + min
	return iInt64
}

func CheckEmail(email string) (bool, error) {
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	matched, err := regexp.MatchString(pattern, email)
	if err != nil {
		return false, err
	}
	if matched {
		return true, nil
	} else {
		return false, nil
	}
}

func CheckPassword(password string) (bool, error) {

	// 大写开头  大小写+数字+特殊字符   8位-16位
	pattern := `^[A-Z][a-zA-Z0-9!@#$%^&*()_+]{7,15}$`
	matched, err := regexp.MatchString(pattern, password)
	if err != nil {
		return false, nil
	}
	return matched, nil
}

// EncryptionEmail 加密邮箱
func EncryptionEmail(email string) string {

	re := regexp.MustCompile(`^([a-zA-Z0-9]+)(@[a-zA-Z0-9]+\.[a-zA-Z]{2,})$`)
	matched := re.FindStringSubmatch(email)
	var newEmail = email
	if len(matched) > 0 {
		username := matched[1]
		usernameLen := len(username)
		var usernameSlice = usernameLen
		if usernameLen > 3 {
			usernameSlice = 3
		}
		domain := matched[2]
		newEmail = fmt.Sprintf("%s***%s", username[:usernameSlice], domain)
	}
	return newEmail
}

func GenerateSalt(length int) (string, error) {
	saltBytes := make([]byte, length)
	_, err := rand.Read(saltBytes)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(saltBytes), nil
}

func HashPassword(password, salt string) string {
	hash := hmac.New(sha256.New, []byte(salt))
	hash.Write([]byte(password))
	return hex.EncodeToString(hash.Sum(nil))
}
