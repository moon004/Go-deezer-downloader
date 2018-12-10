package main_test

import (
	"bytes"
	"crypto/aes"
	"crypto/md5"
	"fmt"
	"testing"

	. "github.com/hopesanddreams/go-decrypt-deezer"
)

const (
	md5Origin    = "51afcde9f56a132096c0496cc95eb24b"
	format       = "3"
	id           = "3135556"
	mediaVersion = "5"
	finAnswer    = "9c2ca4649cc23e7905f09324e9fe1d24505a18b97267b56b8deefecb1d62686d2f5a0bea21e1d6dbd9c8f34c691e12dc83cac650c014d41f69d381b0ce749ff5d38c5e89c566677c9cd24555e6c2bc02"
)

func ErrChecker(t *testing.T, ErrMsg string, err error) {
	if err != nil {
		t.Fatalf("%s: %v", ErrMsg, err)
	}
}
func OnErrorChecker(t *testing.T, err *OnError) {
	if err != nil {
		t.Fatalf("%s: %v", err.Message, err.Error)
	}
}

func Equals(t *testing.T, myanswer, expected string) {
	if myanswer != expected {
		t.Errorf("Expected %s but I get %s", expected, myanswer)
	}
}
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
			ErrChecker(t, "Key Error", err)

			ciphertext := make([]byte, len(input))
			mode := NewECBEncrypter(block)
			mode.CryptBlocks(ciphertext, input)
			Equals(t, fmt.Sprintf("%x", ciphertext), tc.out)
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

			Equals(t, out, tc.out)
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

			Equals(t, BFKey, tc.out)
		})
	}
}

func TestGetUrlDownload(t *testing.T) {
	tt := []struct {
		Name     string
		Expected string
		TrackID  string
	}{
		{
			Name:     "Get the correct Download Url",
			Expected: "https://e-cdns-proxy-5.dzcdn.net/mobile/1/9c2ca4649cc23e7905f09324e9fe1d24505a18b97267b56b8deefecb1d62686d2f5a0bea21e1d6dbd9c8f34c691e12dc83cac650c014d41f69d381b0ce749ff5d38c5e89c566677c9cd24555e6c2bc02",
			TrackID:  "3135556",
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			// Test the most important behavior
			client, err := Login()
			OnErrorChecker(t, err)

			downloadURL, _, client, err := GetUrlDownload(tc.TrackID, client)
			OnErrorChecker(t, err)
			Equals(t, downloadURL, tc.Expected)

		})
	}

}
