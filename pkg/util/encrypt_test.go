package util

import (
	"fmt"
	"testing"
)

func TestGenerateRSAKeyPair(t *testing.T) {
	gotPrivateKey, gotPublicKey, err := GenerateRSAKeyPair(512)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("key:")
	fmt.Println(gotPublicKey)
	fmt.Println("private key:")
	fmt.Println(gotPrivateKey)
}

func TestRasEncrypto(t *testing.T) {
	gotPrivateKey, gotPublicKey, err := GenerateRSAKeyPair(512)
	if err != nil {
		t.Fatal(err)
	}

	message := "hello fluteNAS"

	encryptedStr, err := RSAEncrypt(gotPublicKey.String(), []byte(message))
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(string(encryptedStr)) // 0x...

	str, err := RSADecrypt(gotPrivateKey.String(), encryptedStr)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(string(str)) // hello fluteNAS
	if string(str) != message {
		t.Fail()
	}
}
