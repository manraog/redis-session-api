package main

import (
	"log"
	"net/http"
	"encoding/json"
	"fmt"
	"os" 
	"time"
	
	"github.com/gomodule/redigo/redis"
	"github.com/gofrs/uuid"
)

// In memory Database
var users_db = map[string]string{
	"Hugo": "Hugo123",
	"Paco": "Paco123",
	"Luis": "Luis123",
}

// User Model
type Credentials struct {
	Password string `json:"password"`
	Username string `json:"username"`
}

// Session Model
type Session struct {
    SessionID string `json:"sessionID"`
    Expiration int64  `json:"expiration"`
    Origin string `json:"origin"`
}



// Redis
var cache redis.Conn



func main() {
	// Redis session cache
	redisHostname := getEnv("REDIS_HOSTNAME","localhost")
	redisPort := getEnv("REDIS_PORT","6379")
	redisURL := "redis://" + redisHostname + ":" + redisPort
	fmt.Println(redisURL)
	ca, err := redis.DialURL(redisURL)
	if err != nil {
		fmt.Println("Redis not found ...")	
		panic(err)
	}
	cache = ca
	// Resources
	http.HandleFunc("/profile", Profile)
	http.HandleFunc("/login", Login)
	http.HandleFunc("/refresh", Refresh)
	// Start HTTP server
	fmt.Println("Starting API ...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}


func Login(w http.ResponseWriter, r *http.Request) {
	// Validate method
	if r.Method != http.MethodPost {
        ResponseMessage(w,http.StatusBadRequest,"Wrong HTTP method")
        return
	}

	var creds Credentials
	// Validate body
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		ResponseMessage(w,http.StatusBadRequest,"Bad JSON body")
		return
	}

	// Retrieve password
	realPassword, ok := users_db[creds.Username]

	// Is password correct?
	if !ok || realPassword != creds.Password {
		ResponseMessage(w,http.StatusUnauthorized,"Wrong user or password")
		return
	}

	// Return session
	ResponseSession(w, creds.Username)
	
}

func Profile(w http.ResponseWriter, r *http.Request) {
	// Validate method
	if r.Method != http.MethodGet {
		ResponseMessage(w,http.StatusBadRequest,"Wrong HTTP method")
        return
	}
	// Retrieve session id from Header
	sessionID := r.Header.Get("SessionID")
	if sessionID == "" {
		ResponseMessage(w,http.StatusUnauthorized,"You need to login to get a SessionID")
		return
	}

	// Verify if session id exists on Redis
	redisResponse, err := cache.Do("GET", sessionID)
	if err != nil {
		ResponseMessage(w,http.StatusInternalServerError,"Internal Server Error")
		fmt.Println(err)	
		return
	}
	// If not exists then the session id has expired or user is not authenticated
	if redisResponse == nil {
		ResponseMessage(w,http.StatusUnauthorized,"You need to login to get a SessionID")
		return
	}
	// Show user's profile
	user := fmt.Sprintf("%s", redisResponse)
	ResponseMessage(w,http.StatusOK,"Hi " + user + "!")
}


func Refresh(w http.ResponseWriter, r *http.Request) {
	// Validate method
	if r.Method != http.MethodPost {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusBadRequest)
        w.Write([]byte("{\"message\": \"Bad method\"}"))
        return
	}
	// Retrieve session id from Header
	sessionID := r.Header.Get("SessionID")
	if sessionID == "" {
		ResponseMessage(w,http.StatusUnauthorized,"You need to login to get a SessionID")
		return
	}
	// Validate if session exists
	redisResponse, err := cache.Do("GET", sessionID)
	if err != nil {
		ResponseMessage(w,http.StatusInternalServerError,"Internal Server Error")
		fmt.Println(err)	
		return
	}
	if redisResponse == nil {
		ResponseMessage(w,http.StatusUnauthorized,"SessionID expired, please login")
		return
	}

	// Return new session
	ResponseSession(w, fmt.Sprintf("%s",redisResponse))

	// Delete old session from Redis
	_, err = cache.Do("DEL", sessionID)
	if err != nil {
		ResponseMessage(w,http.StatusInternalServerError,"Internal Server Error")
		fmt.Println(err)	
		return
	}
}




func ResponseMessage(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(code)
    w.Write([]byte("{\"message\": \"" + message + "\"}"))

}

func ResponseSession(w http.ResponseWriter, username string) {
	//Token expiration time
	exp := time.Now().Add(120 * time.Second).Unix()

	// Get Origin hostname
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "NoHostName"
	}

	// Create a new session ID
	i, err := uuid.NewV4()
	if err != nil {
		ResponseMessage(w,http.StatusInternalServerError,"Internal Server Error")
		fmt.Println(err)
		return
	}
	sessionID := i.String()

	// Save sessionID on cache with expiring time of 120 seconds
	_, err = cache.Do("SETEX", sessionID, "120", username)
	if err != nil {
		ResponseMessage(w,http.StatusInternalServerError,"Internal Server Error")
		fmt.Println(err)	
		return
	}

	// Return sessionID
	res := &Session{SessionID: sessionID, Expiration: exp, Origin: hostname}
	response, err := json.Marshal(res)
    if err != nil {
    	ResponseMessage(w,http.StatusInternalServerError,"Internal Server Error")
		fmt.Println(err)	
        return
    }
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(response)
}


func getEnv(key, defaultValue string) string {
    value := os.Getenv(key)
    if len(value) == 0 {
        return defaultValue
    }
    return value
}