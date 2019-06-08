package codec

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
)

// aesCodec for AES/CBC/PKCS7Padding
type aesCodec struct {
	Codec
	key []byte
}

func (c *aesCodec) Encode(origin []byte) ([]byte, error) {
	block, err := aes.NewCipher(c.key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	origin = pkcsPadding(origin, blockSize)
	blockMode := cipher.NewCBCEncrypter(block, c.key[:blockSize])
	crypted := make([]byte, len(origin))
	blockMode.CryptBlocks(crypted, origin)
	return crypted, nil
}

func (c *aesCodec) Decode(encoded []byte) ([]byte, error) {
	block, err := aes.NewCipher(c.key)
	if err != nil {
		return nil, err
	}
	blockModel := cipher.NewCBCDecrypter(block, c.key)
	plantText := make([]byte, len(encoded))
	blockModel.CryptBlocks(plantText, encoded)
	plantText = pkcs7UnPadding(plantText)
	return plantText, nil
}

func pkcs7UnPadding(plantText []byte) []byte {
	length := len(plantText)
	unPadding := int(plantText[length-1])
	return plantText[:(length - unPadding)]
}

func pkcsPadding(cipherText []byte, blockSize int) []byte {
	padding := blockSize - len(cipherText)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(cipherText, padText...)
}
