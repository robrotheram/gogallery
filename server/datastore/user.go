package datastore

import (
	"golang.org/x/crypto/bcrypt"
	"log"
	"math/rand"
	"time"
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
	pasword := RandomPassword(8)
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

const charset = "abcdefghijklmnopqrstuvwxyz" +
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var seededRand *rand.Rand = rand.New(
	rand.NewSource(time.Now().UnixNano()))

func StringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func RandomPassword(length int) string {
	return StringWithCharset(length, charset)
}
