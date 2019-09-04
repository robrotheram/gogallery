package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/robrotheram/gogallery/config"
	"github.com/robrotheram/gogallery/datastore"
	"github.com/robrotheram/gogallery/web"
	"github.com/robrotheram/gogallery/worker"
	"github.com/spf13/viper"
	"golang.org/x/crypto/bcrypt"
)

func generatePassword() {
	args := os.Args[1:]
	if len(args) == 2 {
		if args[0] == "generate" {
			hash, err := bcrypt.GenerateFromPassword([]byte(args[1]), bcrypt.MinCost)
			if err != nil {
				log.Println(err)
			}
			fmt.Printf("Your Admin Password is: %s \nPlease Copy it into the config \n", string(hash))
			os.Exit(0)
		} else {
			log.Fatalf("Unsure what the argument [ %s ] did you mean generate", args[0])
		}
	} else if len(args) == 0 {
		return
	} else {
		log.Fatal("Invalid Arguments")
	}

}

var Config *config.Configuration

func main() {
	generatePassword()

	viper.SetConfigName("config")
	viper.AddConfigPath(".")
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

	if _, err := os.Stat(Config.Gallery.Basepath); os.IsNotExist(err) {
		panic("GALLERY DIRECTORY NOT FOUND EXITING!")
	}

	worker.StartWorkers(Config.Server)
	go setUpWatchers(Config.Gallery.Basepath)
	datastore.Cache = datastore.NewDataStore(&Config.Database)

	go func() {
		datastore.ScanPath(Config.Gallery.Basepath, &Config.Gallery)
	}()

	go func() {
		if Config.IG.SyncRate == 0 {
			fmt.Println("Instagram Sync Cancled")
			return
		}
		d := time.Duration(Config.IG.SyncRate) * time.Minute
		for range time.Tick(d) {
			if Config.IG.Enable {
				fmt.Println("Running Task: Instagram Sync")
				datastore.IG.SyncFrom()
			}
		}
	}()

	web.Serve(Config)
}

var watcher *fsnotify.Watcher

// main
func setUpWatchers(path string) {
	// creates a new file watcher
	watcher, _ = fsnotify.NewWatcher()
	defer watcher.Close()
	// starting at the root of the project, walk each file/directory searching for
	// directories
	if err := filepath.Walk(path, watchDir); err != nil {
		fmt.Println("ERROR", err)
	}
	//
	done := make(chan bool)
	//
	go func() {
		for {
			select {
			// watch for events
			case _ = <-watcher.Events:
				fmt.Println("File chnage detected. Scanning files")
				datastore.ScanPath(Config.Gallery.Basepath, &Config.Gallery)
				// watch for errors
			case err := <-watcher.Errors:
				fmt.Println("ERROR", err)
			}
		}
	}()
	<-done
}

// watchDir gets run as a walk func, searching for directories to add watchers to
func watchDir(path string, fi os.FileInfo, err error) error {
	if fi.Mode().IsDir() {
		return watcher.Add(path)
	}
	return nil
}
