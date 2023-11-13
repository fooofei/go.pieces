package pipehttps

import (
	"crypto/rand"
	"os"
	"path/filepath"
)

func GetCurDir() string {
	var fullPath, _ = os.Executable()
	return filepath.Dir(fullPath)
}

func RandString(n int) []byte {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	var b = make([]byte, n)

	var secret = make([]byte, n)
	rand.Read(secret)

	for i := range secret {
		b[i] = letterBytes[int64(secret[i])%int64(len(letterBytes))]
	}
	return b
}

func TestFileExists(filePath string) error {
	var f, err = os.Open(filePath)
	if err != nil {
		return err
	}
	f.Close()
	return nil
}
