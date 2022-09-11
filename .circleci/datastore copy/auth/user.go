package auth

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/robrotheram/gogallery/backend/datastore"
)

var registerUser = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	id := r.Header.Get("id")
	user := datastore.User{}
	a_user := datastore.FindUserByID(id)
	_ = json.NewDecoder(r.Body).Decode(&user)

	if user.Username != "" {
		a_user.Username = user.Username
	}
	if user.Email != "" {
		a_user.Email = user.Email
	}
	if user.Password != "" {
		a_user.Password = datastore.HashAndSalt(user.Password)
	}
	datastore.Cache.DB.Save(&a_user)
	auth := datastore.User{Username: a_user.Username, Email: a_user.Email}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(auth)
})

func authenticateUser(w http.ResponseWriter, r *http.Request) {
	var user = datastore.User{}
	_ = json.NewDecoder(r.Body).Decode(&user)
	authorised_user := datastore.FindUserByUsername(user.Username)

	if (user.Username == authorised_user.Username) && (datastore.ComparePasswords(authorised_user.Password, user.Password)) {
		fmt.Println("AUTHORIZED")
		token, err := getToken(authorised_user.ID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error generating JWT token: " + err.Error()))
		} else {
			w.Header().Set("Authorization", "Bearer "+token)
			w.WriteHeader(http.StatusOK)
			auth := datastore.User{Token: token, Username: authorised_user.Username, Email: authorised_user.Email}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(auth)
		}
	} else {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Name and password do not match"))
		return
	}
}

var regenTokenHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	id := r.Header.Get("id")
	user := datastore.FindUserByID(id)
	if user.ID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Error User Not Found"))
		return
	}
	token, err := getToken(id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error generating JWT token: " + err.Error()))
	} else {
		w.Header().Set("Authorization", "Bearer "+token)
		w.WriteHeader(http.StatusOK)
		auth := datastore.User{Token: token, Username: user.Username, Email: user.Email}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(auth)
	}
})

func InitAuthRoutes(router *mux.Router) *mux.Router {
	router.HandleFunc("/api/admin/login", authenticateUser).Methods("POST")
	router.Handle("/api/admin/auth/update", AuthMiddleware(registerUser)).Methods("POST")
	router.Handle("/api/admin/authorised", AuthMiddleware(regenTokenHandler))
	return router
}
