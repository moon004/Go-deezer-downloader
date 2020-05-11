package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

const (
	// APIUrl is the hidden Deezer API url
	APIUrl = "http://www.deezer.com/ajax/gw-light.php"
	// Mobile api url, used for loginless download
	MobileUrl = "https://api.deezer.com/1.0/gateway.php"
	// Mobile api key
	ApiKey = "4VCYIJUCDLOUELGD1V8WBVYBNVDYOXEWSLLZDONGBBDFVXTZJRXPR29JRLQFO6ZE"
)

func main() {
	id := cfg.ID
	client := &http.Client{}
	downloadURL, FName, client, err := GetUrlDownload(id, client)
	if err != nil {
		log.Fatalf("%s: %v", err.Message, err.Error)
	}
	err = GetAudioFile(downloadURL, id, FName, client)
	if err != nil {
		log.Fatalf("%s and %v", err.Message, err.Error)
	}
}

// GetUrlDownload get the url for the requested track
func GetUrlDownload(id string, client *http.Client) (string, string, *http.Client, *OnError) {
	// fmt.Println("Getting Download url")
	jsonTrack := &TrackData{}
	//APIToken, _ := GetCSRFToken(client)
	SIDToken, _, _ := GetTokens(client)

	jsonPrep := `{"SNG_ID":"` + id + `"}`
	jsonStr := []byte(jsonPrep)
	req, err := newRequest(MobileUrl, "POST", jsonStr)
	if err != nil {
		return "", "", nil, &OnError{err, "Error during GetUrlDownload request"}
	}
	req = addMobileQs(req, SIDToken)

	resp, _ := client.Do(req)
	body, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	debug("Client request Raw Url " + req.URL.String())
	debug("Body of the resp in GetUrlDownload ", string(body))

	err = json.Unmarshal(body, &jsonTrack)
	if err != nil {
		return "", "", nil, &OnError{err, "Error during GetUrlDownload Unmarshalling"}
	}
	FileSize320, _ := jsonTrack.Results.FileSize320.Int64()
	FileSize256, _ := jsonTrack.Results.FileSize256.Int64()
	FileSize128, _ := jsonTrack.Results.FileSize128.Int64()
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
	songID := jsonTrack.Results.ID.String()
	md5Origin := jsonTrack.Results.MD5Origin.String()
	mediaVersion := jsonTrack.Results.MediaVersion.String()
	songTitle := jsonTrack.Results.SngTitle
	artName := jsonTrack.Results.ArtName
	FName := fmt.Sprintf("%s - %s.mp3", songTitle, artName)
	debug("(md5Origin: %v) (songID: %v) (format: %v) (mediaVersion:%v)",
		md5Origin, songID, format, mediaVersion)

	downloadURL, err := DecryptDownload(md5Origin, songID, format, mediaVersion)
	if err != nil {
		return "", "", nil, &OnError{err, "Error Getting DownloadUrl"}
	}
	debug("The Acquired Download Url:%s", downloadURL)
	return downloadURL, FName, client, nil
}

// GetAudioFile gets the audio file from deezer server
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
	debug("GetAudioFile Response:%v", resp)
	err = DecryptMedia(resp.Body, id, FName, resp.ContentLength)
	if err != nil {
		return &OnError{err, "Error during DecryptMedia"}
	}
	defer resp.Body.Close()
	return nil
}
