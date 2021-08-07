package test


import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"testing"
)

// 这个文件演示如何在 go 语言中使用 sha256rsa 算法

// ParseRSAPrivateKeyFromPEM will get a rsa.PrivateKey from PEM string.
func ParseRSAPrivateKeyFromPEM(privateKeyPEM string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(privateKeyPEM))
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing the private key")
	}

	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed x509.ParsePKCS8PrivateKey err: %w", err)
	}
	rsaKey, ok := privateKey.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("cannot cast to *rsa.PrivateKey")
	}
	return rsaKey, nil
}

func ParsePrivateKey(privateKey string) (*rsa.PrivateKey, error) {
	k := "-----BEGIN PRIVATE KEY-----\n" + privateKey + "\n-----END PRIVATE KEY-----"
	return ParseRSAPrivateKeyFromPEM(k)
}

func Encode(k *rsa.PrivateKey, data []byte) ([]byte, error) {
	shaNew := sha256.New()
	shaNew.Write(data)
	hashed := shaNew.Sum(nil)
	return rsa.SignPKCS1v15(rand.Reader, k, crypto.SHA256, hashed)
}

// google 也有一个例子
// https://cloud.google.com/endpoints/docs/openapi/service-account-authentication?hl=zh-cn
// _ "golang.org/x/oauth2/jws"
// 另一个例子
// https://general.support.brightcove.com/developer/create-json-web-token.html
func sha256WithRSA() {
	// 没有 BEGIN PRIVATE KEY 前缀的
	privKey := ""

	fpath := ""

	content, _ := os.ReadFile(fpath)

	priKey, err := ParsePrivateKey(privKey)
	if err != nil {
		panic(err)
	}

	// 暂时不知道如何对等 Java 中的 sha256WithRsa/Pss
	// 也可能是 SignPSS https://studygolang.com/articles/12701
	signature, err := Encode(priKey, content)
	if err != nil {
		panic(err)
	}
	value := base64.StdEncoding.EncodeToString(signature)
	fmt.Printf("%v\n", value)
}

func TestSha256WithRSA(t *testing.T) {
	sha256WithRSA()
}
