package datastore

import (
	"log"

	"github.com/robrotheram/gogallery/config"
	"golang.org/x/crypto/bcrypt"
)

const ADMINID = "00000"

func FindUserByID(id string) User {
	var user User
	Cache.DB.One("ID", id, &user)
	return user
}
func FindUserByUsername(username string) User {
	var user User
	Cache.DB.One("Username", username, &user)
	return user
}

func CreateDefaultUser() {
	pasword := config.RandomPassword(8)
	user := User{
		ID:       ADMINID,
		Username: "admin",
		Password: HashAndSalt(pasword),
		Email:    "admin@admin.com"}
	Cache.DB.Save(&user)
	log.Printf("New Admin account created! \n Account details: \n username: %s \n password: %s \n", user.Username, pasword)
}

func HashAndSalt(plainPwd string) string {
	plainHash := []byte(plainPwd)
	hash, err := bcrypt.GenerateFromPassword(plainHash, bcrypt.MinCost)
	if err != nil {
		log.Println(err)
	}
	return string(hash)
}
func ComparePasswords(hashedPwd string, plainPwd string) bool {
	byteHash := []byte(hashedPwd)
	plainHash := []byte(plainPwd)
	err := bcrypt.CompareHashAndPassword(byteHash, plainHash)
	if err != nil {
		log.Println(err)
		return false
	}
	return true
}
