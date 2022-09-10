package models

import "time"

type Album struct {
	Id          string       `json:"id" storm:"id"`
	Name        string       `json:"name"`
	ModTime     time.Time    `json:"mod_time"`
	Parent      string       `json:"parent"`
	ParenetPath string       `json:"parentPath,omitempty"`
	Path        string       `json:"path,omitempty"`
	ProfileID   string       `json:"profile_image"`
	Images      []Picture    `json:"images"`
	Children    AlbumStrcure `json:"children"`
	GPS         GPS          `json:"gps"`
}

type AlbumStrcure = map[string]Album

func (a *Album) Update(alb Album) {

	if a.Name != alb.Name && alb.Name != "" {
		a.Name = alb.Name
	}
	if a.Parent != alb.Parent && alb.Parent != "" {
		a.Parent = alb.Parent
	}
	if a.ParenetPath != alb.ParenetPath && alb.ParenetPath != "" {
		a.ParenetPath = alb.ParenetPath
	}
	if (a.ProfileID != alb.ProfileID) && (alb.ProfileID != "") {
		a.ProfileID = alb.ProfileID
	}
	if a.Children == nil {
		a.Children = make(map[string]Album)
	}
	if a.Id == "" {
		a.Id = alb.Id
	}
}
