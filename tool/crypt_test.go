package tool

import (
	"fmt"
	"testing"
)

func Test_AES128Encrypt_AND_AES128Decrypt(t *testing.T) {
	expected := "helloworld"
	key := "8dv4byf8b9e6bc1x"
	iv := "xduio1f8a12348u4"
	encrypt, err := AES128Encrypt([]byte(expected), []byte(key), []byte(iv))
	if err != nil {
		t.Error(err)
	}
	decrypt, err := AES128Decrypt(encrypt, []byte(key), []byte(iv))
	if err != nil {
		t.Error(err)
	}
	de := string(decrypt)
	if de != expected {
		t.Errorf("expected: %s, result: %s", expected, de)
	}
	fmt.Printf("Decrypt result: %s\n", de)
}
