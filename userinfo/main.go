package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

// UserDetails represents the structure for user details
type UserDetails struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Age   string `json:"age"`
}

// AuthTokens represents the structure for authentication tokens
var AuthTokens = make(map[string]string)

var (
	// Logger for writing to the console and log file
	logger *log.Logger
	logFile *os.File
)

// UserDetailsHandler handles GET requests for user details
func UserDetailsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET method is allowed", http.StatusMethodNotAllowed)
		logger.Println("Invalid request method:", r.Method)
		return
	}

	// Fetch headers
	username := r.Header.Get("username")
	authKey := r.Header.Get("Authorization")

	// Validate headers
	if username == "" || authKey == "" {
		http.Error(w, "Missing username or auth_key in headers", http.StatusBadRequest)
		logger.Println("Missing username or auth_key in headers")
		return
	}

	// Load auth tokens
	if err := loadAuthTokens("/app/shared_data/authtokens.json"); err != nil {
		http.Error(w, "Error loading auth tokens", http.StatusInternalServerError)
		logger.Println("Error loading auth tokens:", err)
		return
	}

	// Validate auth key
	if expectedAuthKey, ok := AuthTokens[username]; !ok || expectedAuthKey != authKey {
		http.Error(w, "Invalid authentication credentials", http.StatusUnauthorized)
		logger.Printf("Invalid auth credentials for user: %s\n", username)
		return
	}

	// Load user details
	userStore, err := loadUserStore("/app/shared_data/users.json")
	if err != nil {
		http.Error(w, "Error loading user store", http.StatusInternalServerError)
		logger.Println("Error loading user store:", err)
		return
	}

	// Fetch user details
	if userDetails, ok := userStore[username]; ok {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(userDetails)
		logger.Printf("User details fetched for user: %s\n", username)
	} else {
		http.Error(w, "User not found", http.StatusNotFound)
		logger.Printf("User not found: %s\n", username)
	}
}

// UserAddHandler handles POST requests to add a new user
func UserAddHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		logger.Println("Invalid request method:", r.Method)
		return
	}

	var userDetails UserDetails
	// Parse JSON data from the request body
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&userDetails); err != nil {
		http.Error(w, "Error parsing JSON data", http.StatusBadRequest)
		logger.Println("Error parsing JSON data:", err)
		return
	}

	// Validate form input
	if userDetails.Name == "" || userDetails.Email == "" || userDetails.Age == "" {
		http.Error(w, "Missing user details in form submission", http.StatusBadRequest)
		logger.Println("Missing user details in form submission", userDetails.Name, userDetails.Email, userDetails.Age)
		return
	}

	// Load user store
	userStore, err := loadUserStore("/app/shared_data/users.json")
	if err != nil {
		http.Error(w, "Error loading user store", http.StatusInternalServerError)
		logger.Println("Error loading user store:", err)
		return
	}

	// Check if user already exists
	if _, exists := userStore[userDetails.Name]; exists {
		http.Error(w, "User already exists", http.StatusConflict)
		logger.Printf("Attempted to add an existing user: %s\n", userDetails.Name)
		return
	}

	// Add new user
	userStore[userDetails.Name] = userDetails

	if err := saveUserStore("/app/shared_data/users.json", userStore); err != nil {
		http.Error(w, "Error saving user store", http.StatusInternalServerError)
		logger.Println("Error saving user store:", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("User added successfully"))
	logger.Printf("User added successfully: %s\n", userDetails.Name)
}

// loadAuthTokens loads authentication tokens from a JSON file
func loadAuthTokens(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&AuthTokens); err != nil && err.Error() != "EOF" {
		return err
	}

	return nil
}

// loadUserStore loads user details from a JSON file
func loadUserStore(filename string) (map[string]UserDetails, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var userStore map[string]UserDetails
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&userStore); err != nil && err.Error() != "EOF" {
		return nil, err
	}

	return userStore, nil
}

// saveUserStore saves user details to a JSON file
func saveUserStore(filename string, userStore map[string]UserDetails) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(userStore); err != nil {
		return err
	}

	return nil
}
func CheckAndCreateFile(filename string) error {
	// Check if the file exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		// File does not exist, create it
		file, err := os.Create(filename)
		if err != nil {
			return fmt.Errorf("error creating file: %w", err)
		}
		defer file.Close()

		// Write '{}' to the file
		_, err = file.Write([]byte("{}"))
		if err != nil {
			return fmt.Errorf("error writing to file: %w", err)
		}
	}

	return nil
}
func main() {
	// Open log file
	var err error
	logFile, err = os.OpenFile("/logs/userinfo.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		fmt.Println("Error opening log file:", err)
		os.Exit(1)
	}
	defer logFile.Close()
	CheckAndCreateFile("/app/shared_data/users.json")
	CheckAndCreateFile("/app/shared_data/authtokens.json")
	// Set up logger
	logger = log.New(logFile, "", log.Ldate|log.Ltime|log.Lshortfile)

	http.HandleFunc("/userdetails", UserDetailsHandler)
	http.HandleFunc("/useradd", UserAddHandler)

	fmt.Println("User server started at http://userinfo:8083")
	logger.Println("User server started at http://userinfo:8083")
	log.Fatal(http.ListenAndServe(":8083", nil))
}
