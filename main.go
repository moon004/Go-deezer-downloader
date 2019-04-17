package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"time"
)

const (
	// APIUrl is the deezer API
	APIUrl = "http://www.deezer.com/ajax/gw-light.php"
	// LoginURL is the API for deezer login
	LoginURL = "https://www.deezer.com/ajax/action.php"
	// deezer domain for cookie check
	Domain = "https://www.deezer.com"
)

func main() {
	id := cfg.ID
	client, err := Login()
	if err != nil {
		log.Fatalf("%s: %v", err.Message, err.Error)
	}
	downloadURL, FName, client, err := GetUrlDownload(id, client)
	if err != nil {
		log.Fatalf("%s: %v", err.Message, err.Error)
	}

	err = GetAudioFile(downloadURL, id, FName, client)
	if err != nil {
		log.Fatalf("%s and %v", err.Message, err.Error)
	}
}

// Login will login the user with the provided credentials
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

	CookieURL, _ := url.Parse(Domain)
	debug("Cookies in Login %v", client.Jar.Cookies(CookieURL))
	resp.Body.Close()

	form := url.Values{}
	form.Add("type", "login")
	form.Add("mail", cfg.Username)
	form.Add("password", cfg.Password)
	form.Add("checkFormLogin", Deez.Results.CheckFormLogin)
	req, err = newRequest(LoginURL, "POST", form.Encode())
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
		debug("Login Success!! the Cookies are: ", client.Jar.Cookies(CookieURL))
		// Set the Cookie afte login succesfully
		addCookies(client, CookieURL)
		return client, nil
	}
	return nil, &OnError{err,
		"Can't Login, resp status code is" + string(resp.StatusCode)}
}

func addCookies(client *http.Client, CookieURL *url.URL) {
	expire := time.Now().Add(time.Hour * 24 * 180)
	expire.Format("2006-01-02T15:04:05.999Z07:00")
	creation := time.Now().Format("2006-01-02T15:04:05.999Z07:00")
	lastUsed := time.Now().Format("2006-01-02T15:04:05.999Z07:00")
	rawcookie := fmt.Sprintf("arl=%s; expires=%v; %s creation=%v; lastAccessed=%v;",
		cfg.UserToken,
		expire,
		"path=/; domain=deezer.com; max-age=15552000; httponly=true; hostonly=false;",
		creation,
		lastUsed)
	debug("Expire format in addCookies", expire)
	cookies := []*http.Cookie{
		{
			Name:     "arl",
			Value:    cfg.UserToken,
			Expires:  expire,
			MaxAge:   15552000,
			Domain:   ".deezer.com",
			Path:     "/",
			HttpOnly: true,
			Raw:      rawcookie,
		},
	}

	client.Jar.SetCookies(CookieURL, cookies)

}

// GetUrlDownload get the url for the requested track
func GetUrlDownload(id string, client *http.Client) (string, string, *http.Client, *OnError) {
	// fmt.Println("Getting Download url")
	jsonTrack := &DeezTrack{}

	ParsedAPIUrl, _ := url.Parse(Domain)
	APIToken, _ := GetToken(client)

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

	debug("Cookies in the Request", req.Cookies())

	resp, _ := client.Do(req)
	body, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	debug("Client Cookies after response %v", client.Jar.Cookies(ParsedAPIUrl))
	debug("Client request Raw Url " + req.URL.String())
	debug("Body of the resp in GetUrlDownload ", string(body))

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
	debug("GetAudioFile reposene Cookies: %v", resp.Cookies())
	debug("GetAudioFile Response:%v", resp)
	err = DecryptMedia(resp.Body, id, FName, resp.ContentLength)
	if err != nil {
		return &OnError{err, "Error during DecryptMedia"}
	}
	defer resp.Body.Close()
	return nil
}
