package tool

import (
	"fmt"
	"testing"
)

func Test_AESEncrypt_AND_AESDecrypt(t *testing.T) {
	res, err := AESEncrypt([]byte("0f6845ebafd8d8db"), []byte("HelloWorld"))
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("Encrypt result: %s\n", string(res))
	res, err = AESDecrypt([]byte("0f6845ebafd8d8db"), res)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("Decrypt result: %s\n", res)
}
