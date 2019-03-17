package main

import (
	"fmt"
	"github.com/robrotheram/gogallery/config"
	"github.com/robrotheram/gogallery/datastore"
	"github.com/robrotheram/gogallery/web"
	"github.com/robrotheram/gogallery/worker"
	"github.com/spf13/viper"
	"golang.org/x/crypto/bcrypt"
	"log"
	"os"
	"strings"
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

	log.Println(Config.Database)
	worker.StartWorkers(Config.Server)
	datastore.Cache = datastore.NewDataStore(&Config.Database)
	go func() {
		datastore.ScanPath(Config.Gallery.Basepath)
	}()

	web.Serve(Config)
}
