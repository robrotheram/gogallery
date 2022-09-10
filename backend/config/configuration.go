package config

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/spf13/viper"
)

type Configuration struct {
	Server  ServerConfiguration
	About   AboutConfiguration
	Gallery GalleryConfiguration
	Deploy  DeployConfig
}

type ServerConfiguration struct {
	Port  string
	Debug bool
}

type GalleryConfiguration struct {
	Name             string
	Basepath         string
	Destpath         string
	Url              string
	Theme            string
	AlbumBlacklist   []string
	PictureBlacklist []string
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
		log.Fatalf("Error reading config file, %s", err)
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

func (c *Configuration) Save() {
	viper.Set("about", c.About)
	viper.Set("gallery", c.Gallery)
	viper.Set("server", c.Server)
	viper.WriteConfig()
}

func (c *Configuration) PromptSiteName() {
	prompt := promptui.Prompt{Label: "Site Name", Default: c.Gallery.Name}
	result, _ := prompt.Run()
	c.Gallery.Name = result
}

func (c *Configuration) PromptGalleryBasePath() {
	prompt := promptui.Prompt{
		Label:   "Path to your images",
		Default: c.Gallery.Basepath,
		Validate: func(s string) error {
			if !c.FileExists(s) {
				return fmt.Errorf("path %s, does not exits", s)
			}
			return nil
		},
	}
	result, _ := prompt.Run()
	c.Gallery.Basepath = result
}

func (c *Configuration) PromptGalleryDest() {
	prompt := promptui.Prompt{Label: "Path to destination", Default: c.Gallery.Destpath}
	result, _ := prompt.Run()
	c.Gallery.Destpath = result
}

func (c *Configuration) PromptGalleryTheme() {
	prompt := promptui.Prompt{Label: "Theme to use", Default: "./theme/estnor"}
	result, _ := prompt.Run()
	c.Gallery.Theme = result
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
	if !c.FileExists(c.Gallery.Theme) && c.Gallery.Theme != "default" {
		log.Panic("path to theme does not exist")
	}
}
