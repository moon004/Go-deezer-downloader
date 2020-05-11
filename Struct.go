package main

import "encoding/json"

// OnError Includes both the error msg and the error itself
type OnError struct {
	Error   error
	Message string
}

// ResultList The json struct for getting information for login
type ResultList struct {
	DeezToken      string `json:"checkForm,omitempty"`
	CheckFormLogin string `json:"checkFormLogin,omitempty"`
}

// TrackData Json struct for getting the returned json data
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

// DeezStruct Struct for Json Data for Login
type DeezStruct struct {
	Error   []string    `json:"error,omitempty"`
	Results *ResultList `json:"results,omitempty"`
}

// Data Struct for getting track's json data
type Data struct {
	DATA *TrackData `json:"DATA"`
}

// DeezTrack is Entry Point of the json Data
type DeezTrack struct {
	Error   []string `json:"error,omitempty"`
	Results *Data    `json:"results,omitempty"`
}
