package main

import (
	"bytes"
	"crypto/aes"
	"crypto/md5"
	"fmt"
	"testing"
)

const (
	md5Origin    = "51afcde9f56a132096c0496cc95eb24b"
	format       = "3"
	id           = "3135556"
	mediaVersion = "5"
	finAnswer    = "9c2ca4649cc23e7905f09324e9fe1d24505a18b97267b56b8deefecb1d62686d2f5a0bea21e1d6dbd9c8f34c691e12dc83cac650c014d41f69d381b0ce749ff5d38c5e89c566677c9cd24555e6c2bc02"
)

func TestECBCipher(t *testing.T) {
	inPart := bytes.Replace(
		[]byte(
			"df4797bef68491542a6963b58bac773d¤51afcde9f56a132096c0496cc95eb24b¤3¤3135556¤5¤"),
		[]byte("¤"),
		[]byte{164},
		-1,
	)
	tt := []struct {
		name string
		key  []byte
		in   []byte
		out  string
	}{
		{
			name: "Check Cipher output",
			key:  []byte("jo6aey6haid2Teih"),
			in:   []byte("abc"),
			out:  "90a14c0c4d0104397e5590415b7be2db", // answer got from nodejs
		},
		{
			name: "Check Cipher output",
			key:  []byte("jo6aey6haid2Teih"),
			in:   []byte("hellow World!"),
			out:  "598ad60562de8421a8223645a44e733e", // asnwer got from nodejs
		},
		{
			name: "Check Cipher output",
			key:  []byte("jo6aey6haid2Teih"),
			in:   inPart,
			out:  finAnswer, // asnwer got from nodejs
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			input := Pad(tc.in) // Test the Pad function

			if len(input)%aes.BlockSize != 0 {
				t.Error("Input size is not multiple of block size")
			}

			block, err := aes.NewCipher(tc.key)
			if err != nil {
				t.Errorf("Key error:%v", err)
			}

			ciphertext := make([]byte, len(input))
			mode := NewECBEncrypter(block)
			mode.CryptBlocks(ciphertext, input)
			if tc.out != fmt.Sprintf("%x", ciphertext) {
				t.Errorf("Expected output %s but I get %x", tc.out, ciphertext)
			}
		})
	}
}

func TestMD5(t *testing.T) {
	x := bytes.Replace(
		[]byte("51afcde9f56a132096c0496cc95eb24b¤3¤3135556¤5"), []byte("¤"), []byte{164}, -1)
	tt := []struct {
		name string
		in   []byte
		out  string
	}{
		{
			name: "Check md5 Output",
			in:   []byte("abc"),
			out:  "900150983cd24fb0d6963f7d28e17f72", // answer got from nodejs
		},
		{
			name: "Check md5 Output",
			in:   []byte("hellow World!asddasdsadsada"),
			out:  "880480d8b39ae846cefe8a60fe80f120", // asnwer got from nodejs
		},
		{
			name: "Check md5 Output",
			in:   x,
			out:  "df4797bef68491542a6963b58bac773d", // asnwer got from nodejs
		},
		{
			name: "Check md5 Output",
			in:   []byte("3135556"),
			out:  "29a15fc70fb278009ab6988ce9a422e8", // asnwer got from nodejs
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {

			md5Sum := md5.Sum(tc.in)
			out := fmt.Sprintf("%x", md5Sum)

			if tc.out != out {
				t.Errorf("expected output %s but get %s", tc.out, out)
			}
		})
	}
}

func TestBlowFish(t *testing.T) {
	tt := []struct {
		name string
		in   string
		out  string
	}{
		{
			name: "Test Blowfish output",
			in:   "3135556",
			out:  "llfk9f,7e%u`<d49", // asnwer got from nodejs
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			BFKey := GetBlowFishKey(tc.in)

			if BFKey != tc.out {
				t.Errorf("expected: %s but get: %s", tc.out, BFKey)
			}
		})
	}
}
