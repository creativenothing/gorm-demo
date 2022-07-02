package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var addr = flag.String("addr", ":8080", "http service address")

type User struct {
	Username string `json:"username"`
	Password string `json:password`
}

// respondJSON makes the response with payload as json format
func respondJSON(w http.ResponseWriter, status int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	fmt.Printf("RESPONSE INSIDE RESPONDJSON: %+v\n", response)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write([]byte(response))
}

// respondError makes the error response with payload as json format
func respondError(w http.ResponseWriter, code int, message string) {
	respondJSON(w, code, map[string]string{"error": message})
}

func home(w http.ResponseWriter, r *http.Request) {

	log.Println(r.URL)
	if r.URL.Path != "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	http.ServeFile(w, r, "index.html")
}

func Connect() *gorm.DB {
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		panic(err.Error())
	}
	crud(db)
	return db
}

func crud(db *gorm.DB) {
	db.AutoMigrate(&User{})
	user := User{Username: "admin", Password: "admin"}
	db.Create(&user)
}

func getUserOr404(username string, w http.ResponseWriter, r *http.Request) *User {
	user := User{}
	if err := db.First(&user, User{Username: username}).Error; err != nil {
		return nil
	}
	return &user
}

func login(w http.ResponseWriter, r *http.Request) {

	if r.Method == "OPTIONS" {
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var u User

	err := json.NewDecoder(r.Body).Decode(&u)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	fmt.Printf("/login %s %+v\n", r.Method, u)

	result := getUserOr404(u.Username, w, r)

	if result == nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, result)

	fmt.Printf("RESULT: %+v\n", result)

}

var db = Connect()

func main() {
	http.HandleFunc("/", home)
	http.HandleFunc("/login", login)
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
	fmt.Printf("server running")
}
