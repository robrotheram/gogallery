package datastore

import (
	"fmt"
	"log"

	"github.com/robrotheram/gogallery/backend/config"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID       string `json:"id,omitempty" storm:"id"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	Email    string `json:"email,omitempty"`
	Token    string `json:"token,omitempty"`
}

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
	fmt.Printf("Account details: \n username: %s \n password: %s \n", user.Username, pasword)
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
