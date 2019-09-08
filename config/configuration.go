package config

type Configuration struct {
	Server   ServerConfiguration
	Database DatabaseConfiguration
	About    AboutConfiguration
	Gallery  GalleryConfiguration
	Admin    AdminConfiguration
	IG       InstagramConfiguration
}

type DatabaseConfiguration struct {
	Baseurl string
}

type ServerConfiguration struct {
	Port  string
	Debug bool
}

type InstagramConfiguration struct {
	Username string
	Password string
	Enable   bool
	SyncRate int
}

type GalleryConfiguration struct {
	Name             string
	Basepath         string
	Url              string
	Theme            string
	ImagesPerPage    int
	QueThreshold     int
	AlbumBlacklist   []string
	PictureBlacklist []string
	Renderer         string
}

type AboutConfiguration struct {
	Enable          bool
	Twitter         string
	Facebook        string
	Email           string
	Instagram       string
	Description     string
	Footer          string
	Photographer    string
	ProfilePhoto    string
	BackgroundPhoto string
	Blog            string
	Website         string
}

type AdminConfiguration struct {
	Enable   bool
	Username string
	Password string
}
