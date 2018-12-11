// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Electronic Code Book (ECB) mode.

// ECB provides confidentiality by assigning a fixed ciphertext block to each
// plaintext block.

// See NIST SP 800-38A, pp 08-09
package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"errors"

	"golang.org/x/crypto/blowfish"
)

type ecb struct {
	b         cipher.Block
	blockSize int
}

func newECB(b cipher.Block) *ecb {
	return &ecb{
		b:         b,
		blockSize: b.BlockSize(),
	}
}

type ecbEncrypter ecb

// NewECBEncrypter is the encrypter used by deezer
func NewECBEncrypter(b cipher.Block) cipher.BlockMode {
	return (*ecbEncrypter)(newECB(b))
}

func (x *ecbEncrypter) BlockSize() int { return x.blockSize }

func (x *ecbEncrypter) CryptBlocks(dst, src []byte) {
	if len(src)%x.blockSize != 0 {
		panic("crypto/cipher: input not full blocks")
	}
	if len(dst) < len(src) {
		panic("crypto/cipher: output smaller than input")
	}
	for len(src) > 0 {
		x.b.Encrypt(dst, src[:x.blockSize])
		src = src[x.blockSize:]
		dst = dst[x.blockSize:]
	}
}

type ecbDecrypter ecb

// NewECBDecrypter is used to decrypt the encrypted media
func NewECBDecrypter(b cipher.Block) cipher.BlockMode {
	return (*ecbDecrypter)(newECB(b))
}

func (x *ecbDecrypter) BlockSize() int { return x.blockSize }

func (x *ecbDecrypter) CryptBlocks(dst, src []byte) {
	if len(src)%x.blockSize != 0 {
		panic("crypto/cipher: input not full blocks")
	}
	if len(dst) < len(src) {
		panic("crypto/cipher: output smaller than input")
	}
	for len(src) > 0 {
		x.b.Decrypt(dst, src[:x.blockSize])
		src = src[x.blockSize:]
		dst = dst[x.blockSize:]
	}
}

// Pad is to pad the ciphertext so that it becomes the multiple of 8
func Pad(src []byte) []byte {
	padding := aes.BlockSize - len(src)%aes.BlockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(src, padtext...)
}

// BFDecrypt is the blowfish decrpytion
func BFDecrypt(buf []byte, bfKey string) ([]byte, error) {
	decrypter, err := blowfish.NewCipher([]byte(bfKey)) // 8bytes
	if err != nil {
		return nil, err
	}

	IV := []byte{0, 1, 2, 3, 4, 5, 6, 7} //8 bytes
	if len(buf)%blowfish.BlockSize != 0 {
		return nil, errors.New("The Buf is not a multiple of 8")
	}
	cbcDecrypter := cipher.NewCBCDecrypter(decrypter, IV)
	cbcDecrypter.CryptBlocks(buf, buf)
	return buf, nil
}

// func main() {
// 	key := []byte("jo6aey6haid2Teih")
// 	plaintext := Pad([]byte("abc"))

// 	if len(plaintext)%aes.BlockSize != 0 {
// 		panic("plaintext is not a multiple of block size")
// 	}

// 	block, err := aes.NewCipher(key)
// 	if err != nil {
// 		panic(err)
// 	}

// 	ciphertext := make([]byte, len(plaintext))
// 	mode := NewECBEncrypter(block)
// 	mode.CryptBlocks(ciphertext, plaintext)
// 	fmt.Printf("%x\n", ciphertext)
// }
