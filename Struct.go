package main

import "encoding/json"

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
