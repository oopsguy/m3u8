package codec

import (
	"fmt"
	"testing"
)

func TestAesCodec_Encode_And_Decode(t *testing.T) {
	c, err := NewCodec(AES, []byte("0f6845ebafd8d8db"))
	if err != nil {
		t.Error(err)
	}
	res, err := c.Encode([]byte("HelloWorld"))
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("Encode result: %s\n", string(res))
	res, err = c.Decode(res)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("Decode result: %s\n", res)
}
