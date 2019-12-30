package main

import (
	"bytes"
	"crypto/aes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"

	"github.com/pkg/errors"
)

func newRequest(enPoint, method string, bodyEntity interface{}) (*http.Request, error) {
	var req *http.Request
	var err error
	switch val := bodyEntity.(type) {
	case []byte:
		req, err = http.NewRequest(method, enPoint, bytes.NewBuffer(val))
	case string:
		req, err = http.NewRequest(method, enPoint, strings.NewReader(val))
	default:
		req, err = http.NewRequest(method, enPoint, nil)
	}
	if bodyEntity == nil {
		req, err = http.NewRequest(method, enPoint, nil)
	} else {

	}
	if err != nil {
		return nil, err
	}

	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/62.0.3202.75 Safari/537.36")
	req.Header.Add("Content-Language", "en-US")
	req.Header.Add("Cache-Control", "max-age=0")
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Accept-Charset", "utf-8,ISO-8859-1;q=0.7,*;q=0.3")
	req.Header.Add("Accept-Language", "de-DE,de;q=0.8,en-US;q=0.6,en;q=0.4")
	req.Header.Add("Content-type", "application/json")

	return req, nil
}

func addQs(req *http.Request, args ...string) *http.Request {
	qs := url.Values{}
	qs.Add("api_version", "1.0")
	qs.Add("api_token", args[0]) //args[0] always token
	qs.Add("input", "3")
	qs.Add("method", args[1]) //args[1] always method

	req.URL.RawQuery = qs.Encode()

	return req
}

func addMobileQs(req *http.Request, SIDToken string) *http.Request {
	qs := url.Values{}
	qs.Add("api_key", ApiKey)
	qs.Add("sid", SIDToken)
	qs.Add("output", "3")
	qs.Add("input", "3")
	qs.Add("method", "song_getData")

	req.URL.RawQuery = qs.Encode()

	return req
}

// GetBlowFishKey get the BlowFishkey for decryption
func GetBlowFishKey(id string) string {
	Secret := "g4el58wc0zvf9na1"
	md5Sum := md5.Sum([]byte(id))
	idM5 := fmt.Sprintf("%x", md5Sum)

	var BFKey string
	for i := 0; i < 16; i++ {
		BFKey += fmt.Sprintf("%s", string(idM5[i]^idM5[i+16]^Secret[i]))
	}

	return BFKey
}

// GetTokens Get the required tokens used in api calls
func GetTokens(client *http.Client) (string, string, *OnError) {
	tokens := &Tokens{}
	args := []string{"null", "deezer.getUserData"}
	reqs, err := newRequest(APIUrl, "GET", nil)
	if err != nil {
		return "", "", &OnError{err, "Error during GetToken GET request"}
	}
	reqs = addQs(reqs, args...)
	resp, err := client.Do(reqs)
	if err != nil {
		return "", "", &OnError{err, "Error during GetToken response"}
	}

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	err = json.Unmarshal(body, &tokens)
	if err != nil {
		return "", "", &OnError{err, "Error During Unmarshal"}
	}
	CSRFToken := tokens.Results.DeezToken
	SIDToken := tokens.Results.SessionId

	debug("The CSRF token is %s", CSRFToken)
	debug("The SID token is %s", SIDToken)
	return SIDToken, CSRFToken, nil
}

// DecryptDownload Get the encrypted download link
func DecryptDownload(md5Origin, songID, format, mediaVersion string) (string, error) {
	urlPart := md5Origin + "¤" + format + "¤" + songID + "¤" + mediaVersion
	data := bytes.Replace([]byte(urlPart), []byte("¤"), []byte{164}, -1)
	md5SumVal := fmt.Sprintf("%x", md5.Sum(data))
	urlPart = md5SumVal + "¤" + urlPart + "¤"

	// Encrypt urlPart in hex format
	key := []byte("jo6aey6haid2Teih")
	plaintext := Pad(bytes.Replace([]byte(urlPart), []byte("¤"), []byte{164}, -1))
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	encryptText := make([]byte, len(plaintext))
	mode := NewECBEncrypter(block) // return ECB encryptor
	mode.CryptBlocks(encryptText, plaintext)
	return "https://e-cdns-proxy-" + md5Origin[:1] + ".dzcdn.net/mobile/1/" + fmt.Sprintf("%x", encryptText),
		nil
}

// DecryptMedia decrypts the encrypted media that is returned by Deezer's server
func DecryptMedia(stream io.Reader, id, FName string, streamLen int64) error {
	// fmt.Println("Gopher is decrypting the media file")
	var wg sync.WaitGroup
	chunkSize := 2048
	bfKey := GetBlowFishKey(id)
	errc := make(chan error)
	var err error
	var destBuffer bytes.Buffer // final Product
	debug("resp Body Size: %v", streamLen)
	for position, i := 0, 0; position < int(streamLen); position, i = position+chunkSize, i+1 {
		func(i, position int, streamLen int64, stream io.Reader) {
			debug("Loop %v started", i)
			var chunkString []byte
			// check if stream is of 2048
			if (int(streamLen) - position) >= 2048 {
				chunkSize = 2048
			} else {
				chunkSize = int(streamLen) - position
			}
			buf := make([]byte, chunkSize) // The "chunk" of data
			if _, err = io.ReadFull(stream, buf); err != nil {
				errc <- errors.Wrapf(err, "error at loop %v", i)
			}
			if i%3 > 0 || chunkSize < 2048 {
				chunkString = buf
			} else { //Decrypt and then write to destBuffer
				debug("decrypting at loop: %v", i)
				chunkString, err = BFDecrypt(buf, bfKey)
				if err != nil {
					errc <- errors.Wrapf(err, "error at loop %v", i)
				}
			}
			if _, err := destBuffer.Write(chunkString); err != nil {
				errc <- errors.Wrapf(err, "error at loop %v", i)
			}
			debug("Current DecyptMedia byte: %v/%v loop: %v chunkSize: %v", position, int(streamLen), i, chunkSize)
		}(i, position, streamLen, stream)
	}
	for {
		select {
		case err = <-errc:
			debug("Got Error")
			return err
		default:
			debug("Default")
			wg.Wait()
			debug("FName", FName)
			NameWithoutSlash := strings.ReplaceAll(FName, "/", "∕")
			debug("NameWithoutSlash ", NameWithoutSlash)
			out, err := os.Create(NameWithoutSlash)
			if err != nil {
				return err
			}
			length, err := destBuffer.WriteTo(out) // You might change from destBuffer.WriteTo(out) to destBuffer.WriteTo(os.Stdout)
			if err != nil {
				return err
			}
			debug("Size Written: %v", length)

			return nil
		}
	}

}
