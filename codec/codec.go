package codec

import (
	"errors"
	"fmt"
	"strings"
)

const (
	AES = "AES"
)

type Codec interface {
	Encode([]byte) ([]byte, error)
	Decode([]byte) ([]byte, error)
}

func NewCodec(codecType string, key []byte) (Codec, error) {
	switch strings.ToUpper(codecType) {
	case AES:
		return &aesCodec{
			key: key,
		}, nil
	default:
		return nil, errors.New(fmt.Sprintf("unsupported encryption: %s", codecType))
	}
}
