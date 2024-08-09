package util

import (
	crand "crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"strings"
)

var alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

func RandStringRunes(n int) string {
	sb := strings.Builder{}
	sb.Grow(n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, rand.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = rand.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(alphabet) {
			sb.WriteByte(alphabet[idx])
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return sb.String()
}

// // 生成一对密钥
// func GenAccessKeySecret() (accessKey string, accessSecret string) {
// 	accessKey = RandStringRunes(16)
// 	accessSecret = RandStringRunes(32)
// 	return
// }

// 生成一对非对称加密的公钥和私钥
func GenerateRSAKeyPair(bits int) (privateKey *LinkedRune, publicKey *LinkedRune, err error) {
	// 生成私钥
	privateKeyObj, err := rsa.GenerateKey(crand.Reader, bits)
	if err != nil {
		return nil, nil, err
	}

	// 编码私钥为PEM格式
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKeyObj)
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	})

	// 生成公钥
	publicKeyObj := &privateKeyObj.PublicKey

	// 编码公钥为PEM格式
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(publicKeyObj)
	if err != nil {
		return nil, nil, err
	}
	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: publicKeyBytes,
	})

	return NewLinkedRune(string(privateKeyPEM)), NewLinkedRune(string(publicKeyPEM)), nil
}

func ParseRsaPublicKey(publicKey string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(publicKey))
	if block == nil {
		return nil, errors.New("failed to parse public PEM block")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %v", err)
	}

	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("convert to rsa.PublicKey failed")
	}
	return rsaPub, nil
}

func ParseRasPrivateKey(privateKey string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(privateKey))
	if block == nil {
		return nil, errors.New("failed to parse private PEM block")
	}

	privateKeyObj, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %v", err)
	}

	return privateKeyObj, nil
}

func RSAEncrypt(publicKey string, message []byte) ([]byte, error) {
	rsaPub, err := ParseRsaPublicKey(publicKey)
	if err != nil {
		return nil, err
	}
	// 使用公钥加密
	encryptedMessage, err := rsa.EncryptPKCS1v15(crand.Reader, rsaPub, message)
	if err != nil {
		log.Fatalf("ras encrypt message failed: %v", err)
	}

	return encryptedMessage, nil
}

func RSADecrypt(privateKey string, encryptedMessage []byte) (string, error) {
	rsaPrivate, err := ParseRasPrivateKey(privateKey)
	if err != nil {
		return "", err
	}

	// 使用私钥解密
	decryptedMessage, err := rsa.DecryptPKCS1v15(crand.Reader, rsaPrivate, encryptedMessage)
	if err != nil {
		log.Fatalf("decrypt message failed: %v", err)
	}

	return string(decryptedMessage), nil
}
