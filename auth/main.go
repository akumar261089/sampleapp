package main

import (
	"encoding/json"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"os"
)

// User represents a simple user structure
type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// TokenResponse represents the response with the auth token
type TokenResponse struct {
	AuthToken string `json:"auth_token"`
	Message   string `json:"message"`
}

var (
	// Logger for writing to the console and log file
	logger *log.Logger
	logFile *os.File
)

// AuthHandler handles authentication requests
func AuthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		logger.Println("Invalid request method:", r.Method)
		return
	}

	// Load userStore from file for each request
	userStore, err := LoadUserStore("../userdata/users.json")
	if err != nil {
		http.Error(w, "Error loading user store", http.StatusInternalServerError)
		logger.Println("Error loading user store:", err)
		return
	}

	var user User
	err = json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		logger.Println("Error decoding request body:", err)
		return
	}

	// Check if the username exists and the password matches the username
	if _, exists := userStore[user.Username]; exists && user.Password == user.Username {
		authToken := generateAuthToken()
		response := TokenResponse{
			AuthToken: authToken,
			Message:   "Authentication successful",
		}

		// Update auth tokens file
		if err := UpdateAuthTokens(user.Username, authToken); err != nil {
			http.Error(w, "Error updating auth tokens", http.StatusInternalServerError)
			logger.Println("Error updating auth tokens:", err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		logger.Println("Authentication successful for user:", user.Username)
	} else {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		logger.Println("Invalid credentials for user:", user.Username)
	}
}

// generateAuthToken generates a simple UUID-based token
func generateAuthToken() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		logger.Println("Error generating auth token:", err)
		return ""
	}
	return hex.EncodeToString(b)
}

// LoadUserStore loads user details from a JSON file
func LoadUserStore(filename string) (map[string]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var userStore map[string]map[string]string
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&userStore); err != nil {
		return nil, err
	}

	// Flatten the nested map
	flatUserStore := make(map[string]string)
	for username, details := range userStore {
		if password, exists := details["name"]; exists {
			flatUserStore[username] = password
		}
	}

	return flatUserStore, nil
}

// UpdateAuthTokens updates the auth tokens file with a new token for the user
func UpdateAuthTokens(username, authToken string) error {
	// Load existing auth tokens
	file, err := os.OpenFile("../userdata/authtokens.json", os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	var authTokens map[string]string
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&authTokens); err != nil && err.Error() != "EOF" {
		return err
	}

	// Update or add the new token
	if authTokens == nil {
		authTokens = make(map[string]string)
	}
	authTokens[username] = authToken

	// Rewind file and write updated tokens
	file.Seek(0, 0) // Move to the beginning of the file
	file.Truncate(0) // Clear the file
	encoder := json.NewEncoder(file)
	if err := encoder.Encode(authTokens); err != nil {
		return err
	}

	return nil
}

func main() {
	// Open log file
	var err error
	logFile, err = os.OpenFile("auth_server.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		fmt.Println("Error opening log file:", err)
		os.Exit(1)
	}
	defer logFile.Close()

	// Set up logger
	logger = log.New(logFile, "", log.Ldate|log.Ltime|log.Lshortfile)

	http.HandleFunc("/auth", AuthHandler)

	fmt.Println("Authentication server started at http://localhost:8082")
	logger.Println("Authentication server started at http://localhost:8082")
	log.Fatal(http.ListenAndServe(":8082", nil))
}
