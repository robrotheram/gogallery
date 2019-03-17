package config

type Configuration struct {
	Server   ServerConfiguration
	Database DatabaseConfiguration
	About    AboutConfiguration
	Gallery  GalleryConfiguration
	Admin    AdminConfiguration
}

type DatabaseConfiguration struct {
	Baseurl string
}

type ServerConfiguration struct {
	Port    string
	Workers int
}

type GalleryConfiguration struct {
	Name          string
	Basepath      string
	Url           string
	Theme         string
	ImagesPerPage int
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
