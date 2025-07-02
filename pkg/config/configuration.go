package config

import (
	"log"
	"os"
	"strings"

	"github.com/spf13/viper"
)

type Configuration struct {
	UI      UIConfiguration
	About   AboutConfiguration
	Gallery GalleryConfiguration
	Deploy  DeployConfig
}

type UIConfiguration struct {
	Public        bool
	Notification  bool
	Theme         string
	ImagesPerPage int
}

type GalleryConfiguration struct {
	Name             string
	Basepath         string
	Destpath         string
	ImagesPerPage    int
	Url              string
	Theme            string
	AlbumBlacklist   []string
	PictureBlacklist []string
	UseOriginal      bool
}

type AboutConfiguration struct {
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
	Github          string
}

type DeployConfig struct {
	SiteId    string
	Draft     bool
	AuthToken string
}

var Config = &Configuration{}

func LoadConfig() *Configuration {
	viper.SetEnvPrefix("GLLRY")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("config not found creating default")
		DefaultConfig()
	}
	err := viper.Unmarshal(&Config)
	if err != nil {
		log.Fatalf("unable to decode into struct, %v", err)
	}
	return Config
}
func (c *AboutConfiguration) Save() {
	log.Println("Saving About Config")
	viper.Set("about", c)
	viper.WriteConfig()
	Config.About = *c
}
func (c *GalleryConfiguration) Save() {
	log.Println("Saving Gallery Config")
	viper.Set("gallery", c)
	viper.WriteConfig()
	Config.Gallery = *c
}
func (c *DeployConfig) Save() {
	log.Println("Saving Deployment Config")
	viper.Set("deploy", c)
	viper.WriteConfig()
	Config.Deploy = *c
}
func (c *Configuration) Save() {
	viper.Set("about", c.About)
	viper.Set("gallery", c.Gallery)
	viper.Set("ui", c.UI)
	viper.Set("deploy", c.Deploy)
	viper.WriteConfig()
}

func (c *Configuration) FileExists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

func (c *Configuration) Validate() {
	if !c.FileExists(c.Gallery.Basepath) {
		log.Panic("path to images does not exist")
		os.Exit(1)
	}
}

func DefaultConfig() {
	Config = &Configuration{
		UI:    UIConfiguration{},
		About: AboutConfiguration{},
		Gallery: GalleryConfiguration{
			Theme: "default",
		},
	}
	// Ensure config file is created if it doesn't exist
	if !Config.FileExists(viper.ConfigFileUsed()) {
		viper.SetConfigName(".gogallery")
		viper.SetConfigType("yaml")
		_ = viper.SafeWriteConfig() // ignore error if file exists
	}
	Config.Save()
}
