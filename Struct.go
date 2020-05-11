package main

import "encoding/json"

// OnError Includes both the error msg and the error itself
type OnError struct {
	Error   error
	Message string
}

// The json struct for getting the CSRF and SID tokens
type Tokens struct {
	Error   []string `json:"error,omitempty"`
	Results struct {
		DeezToken      string `json:"checkForm"`
		SessionId      string `json:"SESSION_ID"`
	} `json:"results"`
}

// TrackData Json struct for getting the returned json data
type TrackData struct {
	Error   []string `json:"error,omitempty"`
	Results struct {
		ID           json.Number `json:"SNG_ID"`
		MD5Origin    json.Number `json:"PUID"`
		FileSize320  json.Number `json:"FILESIZE_MP3_320"`
		FileSize256  json.Number `json:"FILESIZE_MP3_256"`
		FileSize128  json.Number `json:"FILESIZE_MP3_128"`
		MediaVersion json.Number `json:"MEDIA_VERSION"`
		SngTitle     string      `json:"SNG_TITLE"`
		ArtName      string      `json:"ART_NAME"`
	} `json:"results"`
}
