package config

import (
	"errors"
	"fmt"
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
	GeminiApiKey  string
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

const DefaultTheme = "EmeraldNoir"

var Config = &Configuration{}

func LoadConfig() (*Configuration, error) {
	viper.SetConfigPermissions(0o600)
	viper.SetEnvPrefix("GLLRY")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		var notFound viper.ConfigFileNotFoundError
		if !errors.As(err, &notFound) && !errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("read configuration: %w", err)
		}
		log.Printf("config not found; creating default")
		if err := DefaultConfig(); err != nil {
			return nil, err
		}
	} else if err := secureConfigPermissions(); err != nil {
		return nil, err
	}
	err := viper.Unmarshal(Config)
	if err != nil {
		return nil, fmt.Errorf("decode configuration: %w", err)
	}
	Config.Gallery.Theme = NormalizeTheme(Config.Gallery.Theme)
	return Config, nil
}

func NormalizeTheme(theme string) string {
	theme = strings.TrimSpace(theme)
	if theme == "" || strings.EqualFold(theme, "default") {
		return DefaultTheme
	}
	return theme
}

func (c *AboutConfiguration) Save() error {
	log.Println("Saving About Config")
	viper.Set("about", c)
	if err := writeConfigSecurely(); err != nil {
		return fmt.Errorf("save author settings: %w", err)
	}
	Config.About = *c
	return nil
}
func (c *GalleryConfiguration) Save() error {
	log.Println("Saving Gallery Config")
	viper.Set("gallery", c)
	if err := writeConfigSecurely(); err != nil {
		return fmt.Errorf("save gallery settings: %w", err)
	}
	Config.Gallery = *c
	return nil
}
func (c *DeployConfig) Save() error {
	log.Println("Saving Deployment Config")
	viper.Set("deploy", c)
	if err := writeConfigSecurely(); err != nil {
		return fmt.Errorf("save deployment settings: %w", err)
	}
	Config.Deploy = *c
	return nil
}
func (c *Configuration) Save() error {
	setViperConfig(c)
	if err := writeConfigSecurely(); err != nil {
		return fmt.Errorf("save configuration: %w", err)
	}
	*Config = *c
	return nil
}

func setViperConfig(c *Configuration) {
	viper.Set("about", c.About)
	viper.Set("gallery", c.Gallery)
	viper.Set("ui", c.UI)
	viper.Set("deploy", c.Deploy)
}

func writeConfigSecurely() error {
	viper.SetConfigPermissions(0o600)
	if err := viper.WriteConfig(); err != nil {
		return err
	}
	return secureConfigPermissions()
}

func secureConfigPermissions() error {
	path := viper.ConfigFileUsed()
	if path == "" {
		return fmt.Errorf("configuration file path is unavailable")
	}
	if err := os.Chmod(path, 0o600); err != nil {
		return fmt.Errorf("set permissions on %q: %w", path, err)
	}
	return nil
}

func (c *Configuration) FileExists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

func (c *Configuration) Validate() error {
	basePath := strings.TrimSpace(c.Gallery.Basepath)
	if basePath == "" {
		return fmt.Errorf("gallery base path is not configured")
	}
	info, err := os.Stat(basePath)
	if err != nil {
		return fmt.Errorf("open gallery base path %q: %w", basePath, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("gallery base path %q is not a directory", basePath)
	}
	return nil
}

func DefaultConfig() error {
	viper.SetConfigPermissions(0o600)
	*Config = Configuration{
		UI: UIConfiguration{
			Theme:         "dark",
			ImagesPerPage: 20,
		},
		About: AboutConfiguration{},
		Gallery: GalleryConfiguration{
			Name:          "GoGallery",
			ImagesPerPage: 50,
			Theme:         DefaultTheme,
		},
	}
	setViperConfig(Config)
	// Ensure a discoverable config file is created if no explicit file was set.
	if viper.ConfigFileUsed() == "" {
		viper.SetConfigName(".gogallery")
		viper.SetConfigType("yaml")
		if err := viper.SafeWriteConfig(); err != nil {
			var exists viper.ConfigFileAlreadyExistsError
			if !errors.As(err, &exists) {
				return fmt.Errorf("create default configuration: %w", err)
			}
		}
		if err := viper.ReadInConfig(); err != nil {
			return fmt.Errorf("read default configuration: %w", err)
		}
	}
	if err := Config.Save(); err != nil {
		return err
	}
	return nil
}
