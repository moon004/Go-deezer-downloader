package main

import (
	"bytes"
	"crypto/aes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strconv"
	"strings"
)

const (
	APIUrl   = "http://www.deezer.com/ajax/gw-light.php"
	LoginUrl = "https://www.deezer.com/ajax/action.php"
)

type OnError struct {
	Error   error
	Message string
}

type ResultList struct {
	DeezToken      string `json:"checkForm,omitempty"`
	CheckFormLogin string `json:"checkFormLogin,omitempty"`
}

type TrackData struct {
	ID           json.Number `json:"SNG_ID"`
	MD5Origin    json.Number `json:"MD5_ORIGIN"`
	FileSize320  json.Number `json:"FILESIZE_MP3_320"`
	FileSize256  json.Number `json:"FILESIZE_MP3_256"`
	FileSize128  json.Number `json:"FILESIZE_MP3_128"`
	MediaVersion json.Number `json:"MEDIA_VERSION"`
	SngTitle     string      `json:"SNG_TITLE"`
	ArtName      string      `json:"ART_NAME"`
}

type DeezStruct struct {
	Error   []string    `json:"error,omitempty"`
	Results *ResultList `json:"results,omitempty"`
}

type Data struct {
	DATA *TrackData `json:"DATA"`
}

type DeezTrack struct {
	Error   []string `json:"error,omitempty"`
	Results *Data    `json:"results,omitempty"`
}

func main() {
	// fmt.Println("Program Started")
	id := cfg.ID
	client, err := Login()
	if err != nil {
		log.Fatalf("%s: %v", err.Message, err.Error)
	}
	downloadURL, FName, client, err := GetUrlDownload(id, client)
	if err != nil {
		log.Fatalf("%s: %v", err.Message, err.Error)
	}

	GetAudioFile(downloadURL, id, FName, client)
}

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

func Login() (*http.Client, *OnError) {
	CookieJar, _ := cookiejar.New(nil)
	client := &http.Client{
		Jar: CookieJar,
	}
	Deez := &DeezStruct{}
	req, err := newRequest(APIUrl, "POST", nil)
	args := []string{"null", "deezer.getUserData"}
	req = addQs(req, args...)
	debug("Header of First request in Login: %v", req.Header)
	resp, err := client.Do(req)
	body, _ := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(body, &Deez)
	if err != nil {
		return nil, &OnError{err, "Error during getCheckFormLogin Unmarshalling"}
	}

	CookieURL, _ := url.Parse(APIUrl)
	debug("Cookies in Login %v", client.Jar.Cookies(CookieURL))
	resp.Body.Close()

	form := url.Values{}
	form.Add("type", "login")
	form.Add("mail", cfg.Username)
	form.Add("password", cfg.Password)
	form.Add("checkFormLogin", Deez.Results.CheckFormLogin)
	req, err = newRequest(LoginUrl, "POST", form.Encode())
	if err != nil {
		return nil, &OnError{err, "Error during Login Request"}
	}

	req.Header.Set("Content-type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(form.Encode())))

	debug("The Header of Login Request", req.Header)

	resp, err = client.Do(req)
	if err != nil {
		return nil, &OnError{err, "Error during Login response"}
	}

	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, &OnError{err, "Error During Login response Read Body"}
	}
	debug("The Login response Body %s", string(body))
	if resp.StatusCode == 200 {
		debug("Login Success!!")
		return client, nil
	}
	return nil, &OnError{err,
		"Can't Login, resp status code is" + string(resp.StatusCode)}
}

func GetUrlDownload(id string, client *http.Client) (string, string, *http.Client, *OnError) {
	// fmt.Println("Getting Download url")
	jsonTrack := &DeezTrack{}

	ParsedAPIUrl, _ := url.Parse(APIUrl)
	APIToken, _ := GetToken(client, ParsedAPIUrl)

	jsonPrep := `{"sng_id":"` + id + `"}`
	jsonStr := []byte(jsonPrep)
	req, err := newRequest(APIUrl, "POST", jsonStr)
	if err != nil {
		return "", "", nil, &OnError{err, "Error during GetUrlDownload request"}
	}

	qs := url.Values{}
	qs.Add("api_version", "1.0")
	qs.Add("api_token", APIToken)
	qs.Add("input", "3")
	qs.Add("method", "deezer.pageTrack")
	req.URL.RawQuery = qs.Encode()

	resp, _ := client.Do(req)
	body, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	debug("Client Cookies after response %v", client.Jar.Cookies(ParsedAPIUrl))
	debug("Client request Raw Url" + req.URL.String())

	err = json.Unmarshal(body, &jsonTrack)
	if err != nil {
		return "", "", nil, &OnError{err, "Error during GetUrlDownload Unmarshalling"}
	}
	FileSize320, _ := jsonTrack.Results.DATA.FileSize320.Int64()
	FileSize256, _ := jsonTrack.Results.DATA.FileSize256.Int64()
	FileSize128, _ := jsonTrack.Results.DATA.FileSize128.Int64()
	var format string
	switch {
	case FileSize320 > 0:
		format = "3"
	case FileSize256 > 0:
		format = "5"
	case FileSize128 > 0:
		format = "1"
	default:
		format = "8"
	}
	songID := jsonTrack.Results.DATA.ID.String()
	md5Origin := jsonTrack.Results.DATA.MD5Origin.String()
	mediaVersion := jsonTrack.Results.DATA.MediaVersion.String()
	songTitle := jsonTrack.Results.DATA.SngTitle
	artName := jsonTrack.Results.DATA.ArtName
	FName := fmt.Sprintf("%s - %s.mp3", songTitle, artName)
	debug("(md5Origin: %v) (songID: %v) (format: %v) (mediaVersion:%v)",
		md5Origin, songID, format, mediaVersion)

	downloadUrl, err := DecryptDownload(md5Origin, songID, format, mediaVersion)
	if err != nil {
		return "", "", nil, &OnError{err, "Error Getting DownloadUrl"}
	}
	debug("The Acquired Download Url:%s", downloadUrl)
	return downloadUrl, FName, client, nil
}

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

func GetAudioFile(downloadURL, id, FName string, client *http.Client) *OnError {
	// fmt.Println("Gopher's getting the audio File")
	req, err := newRequest(downloadURL, "GET", nil)
	if err != nil {
		return &OnError{err, "Error during GetAudioFile Get request"}
	}

	resp, err := client.Do(req)
	if err != nil {
		return &OnError{err, "Error during GetAudioFile response"}
	}
	debug("GetAudioFile reposene Cookies: %v", resp.Cookies())
	debug("GetAudioFile Response:%v", resp)
	DecryptMedia(resp.Body, id, FName, resp.ContentLength)
	defer resp.Body.Close()
	return nil
}

func DecryptMedia(stream io.Reader, id, FName string, streamLen int64) error {
	// fmt.Println("Gopher is decrypting the media file")
	chunkSize := 2048
	bfKey := GetBlowFishKey(id)
	i := 0
	position := 0
	var err error
	var destBuffer bytes.Buffer // final Product
	debug("resp Body Size: %v", streamLen)
	for position < int(streamLen) {
		var chunkString []byte
		// check if stream is of 2048
		if (int(streamLen) - position) >= 2048 {
			chunkSize = 2048
		} else {
			chunkSize = int(streamLen) - position
		}
		buf := make([]byte, chunkSize) // The "chunk" of data
		if _, err = io.ReadFull(stream, buf); err != nil {
			return err
		}
		if i%3 > 0 || chunkSize < 2048 {
			chunkString = buf
		} else { //Decrypt and then write to destBuffer
			chunkString, err = BFDecrypt(buf, bfKey)
			if err != nil {
				return err
			}
		}
		if _, err := destBuffer.Write(chunkString); err != nil {
			return err
		}
		position += chunkSize
		i++
		debug("Current DecyptMedia byte: %v", position)
	}
	out, err := os.Create(FName)
	if err != nil {
		return err
	}
	length, err := destBuffer.WriteTo(out) // You might change form destBuffer.WriteTo(out) to destBuffer.WriteTo(os.Stdout)
	if err != nil {
		return err
	}
	debug("Size Written: %v", length)

	return nil
}

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

func GetToken(client *http.Client, ParsedAPIUrl *url.URL) (string, *OnError) {
	Deez := &DeezStruct{}
	args := []string{"null", "deezer.getUserData"}
	reqs, err := newRequest(APIUrl, "GET", nil)
	if err != nil {
		return "", &OnError{err, "Error during GetToken GET request"}
	}
	reqs = addQs(reqs, args...)
	resp, err := client.Do(reqs)
	if err != nil {
		return "", &OnError{err, "Error during GetToken response"}
	}

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	err = json.Unmarshal(body, &Deez)
	if err != nil {
		return "", &OnError{err, "Error During Unmarshal"}
	}

	APIToken := Deez.Results.DeezToken

	debug("Display the Token %s", APIToken)
	return APIToken, nil
}
