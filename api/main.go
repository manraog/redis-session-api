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
	ca, err := redis.DialURL("redis://redis:6379")
	if err != nil {
		panic(err)
		fmt.Println("Redis not found ...")	
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
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusBadRequest)
        w.Write([]byte("{\"message\": \"Bad method\"}"))
        return
	}

	var creds Credentials
	// Validate body
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("{\"message\": \"Bad body\"}"))
		return
	}

	// Retrieve password
	realPassword, ok := users_db[creds.Username]

	// Is password correct?
	if !ok || realPassword != creds.Password {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("{\"message\": \"Wrong user or password\"}"))
		return
	}

	//Token expiration time
	exp := time.Now().Add(180 * time.Second).Unix()

	// Get Origin hostname
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "Error"
	}

	// Create a new session ID
	i, err := uuid.NewV4()
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("{\"message\": \"Internal server error\"}"))
		fmt.Println(err)	
		return
	}
	sessionID := i.String()

	// Save sessionID on cache with expiring time of 180 seconds
	_, err = cache.Do("SETEX", sessionID, "180", creds.Username)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("{\"message\": \"Internal server error\"}"))
		fmt.Println(err)	
		return
	}

	// Return sessionID
	res := &Session{SessionID: sessionID, Expiration: exp, Origin: hostname}
	response, err := json.Marshal(res)
    if err != nil {
        w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("{\"message\": \"Internal server error\"}"))
		fmt.Println(err)	
        return
    }
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(response)
	
}

func Profile(w http.ResponseWriter, r *http.Request) {
	// Validate method
	if r.Method != http.MethodGet {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusBadRequest)
        w.Write([]byte("{\"message\": \"Bad method\"}"))
        return
	}
	// Retrieve session id from Header
	sessionID := r.Header.Get("SessionID")
	if sessionID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("{\"message\": \"You need to login to get a SessionID\"}"))
		return
	}

	// Verify if session id exists on Redis
	redisResponse, err := cache.Do("GET", sessionID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("{\"message\": \"Internal server error\"}"))
		fmt.Println(err)	
		return
	}
	// If not exists then the session id has expired or user is not authenticated
	if redisResponse == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("{\"message\": \"You need to login\"}"))
		return
	}
	// Show user's profile
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("{\"message\": \"Hi %s!\"}", redisResponse)))
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
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("{\"message\": You need to login to get a SessionID}"))
		return
	}
	// Validate if session exists
	redisResponse, err := cache.Do("GET", sessionID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("{\"message\": \"Internal server error\"}"))
		fmt.Println(err)	
		return
	}
	if redisResponse == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("{\"message\": \"SessionID expired, please login\"}"))
		return
	}
	// New sessionID
	u, err := uuid.NewV4()
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("{\"message\": \"Internal server error\"}"))
		fmt.Println(err)	
		return
	}
	newSessionID := u.String()
	// Save new session to Redis
	_, err = cache.Do("SETEX", newSessionID, "180", fmt.Sprintf("%s",redisResponse))
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("{\"message\": \"Internal server error\"}"))
		fmt.Println(err)	
		return
	}
	// Delete old session from Redis
	_, err = cache.Do("DEL", sessionID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("{\"message\": \"Internal server error\"}"))
		fmt.Println(err)	
		return
	}
	// Return new sessionID
	response := map[string]string{"SessionID": newSessionID}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}