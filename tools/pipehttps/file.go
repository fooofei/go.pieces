package main

import (
	"crypto/rand"
	"os"
	"path/filepath"
)

func generateTmpFile() string {
	var fullPath, _ = os.Executable()
	var cur = filepath.Dir(fullPath)
	return filepath.Join(cur, string(randString(16)))
}

func randString(n int) []byte {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	var b = make([]byte, n)

	var secret = make([]byte, n)
	rand.Read(secret)

	for i := range secret {
		b[i] = letterBytes[int64(secret[i])%int64(len(letterBytes))]
	}
	return b
}
