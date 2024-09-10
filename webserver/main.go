package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type TokenResponse struct {
	AuthToken string `json:"auth_token"`
	Message   string `json:"message"`
}

// Product represents the structure for a product item
type Product struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Price string `json:"price"`
}

// User represents the structure for a user
type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// UserDetails represents the user details fetched from the API
type UserDetails struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Age   string    `json:"age"`
}

var tmpl *template.Template
var logFile *os.File

func init() {
	var err error

	// Open or create the log file
	logFile, err = os.OpenFile("/logs/webserver.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("ERROR: Error opening log file: %v\n", err)
		os.Exit(1)
	}

	// Set up logging to file and console
	log.SetOutput(logFile)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	tmpl = template.Must(template.ParseGlob("templates/*"))
}
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET method is allowed", http.StatusMethodNotAllowed)
		log.Println("Invalid request method:", r.Method)
		return
	}

	// Health check response
	response := map[string]string{"status": "ok"}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
	log.Println("Health check requested.")
}
// HomeHandler serves the home page with product details
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("HomeHandler: Fetching product details")

	resp, err := http.Get("http://productlist:8081/products") // Replace with actual API URL
	if err != nil {
		log.Printf("ERROR: HomeHandler: Error fetching products - %v\n", err)
		http.Error(w, "Unable to fetch products", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("ERROR: HomeHandler: Unexpected status code %d from products API\n", resp.StatusCode)
		http.Error(w, "Error fetching products", http.StatusInternalServerError)
		return
	}

	var products []Product
	if err := json.NewDecoder(resp.Body).Decode(&products); err != nil {
		log.Printf("ERROR: HomeHandler: Error parsing product data - %v\n", err)
		http.Error(w, "Error parsing product data", http.StatusInternalServerError)
		return
	}

	log.Printf("HomeHandler: Successfully fetched %d products\n", len(products))
	tmpl.ExecuteTemplate(w, "home.html", products)
}

// LoginHandler serves the login page and handles login form submission
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		http.SetCookie(w, &http.Cookie{
			Name:     "auth_key",
			Value:    "",
			Path:     "/",
			HttpOnly: true,
			Expires:  time.Now().Add(-24 * time.Hour), // Set to a past date to delete
		})
		http.SetCookie(w, &http.Cookie{
			Name:     "username",
			Value:    "",
			Path:     "/",
			HttpOnly: true,
			Expires:  time.Now().Add(-24 * time.Hour), // Set to a past date to delete
		})
		username := r.FormValue("username")
		password := r.FormValue("password")

		user := User{Username: username, Password: password}
		userJson, _ := json.Marshal(user)

		log.Printf("LoginHandler: Attempting to log in user %s\n", username)

		resp, err := http.Post("http://auth:8082/auth", "application/json", strings.NewReader(string(userJson)))
		if err != nil {
			log.Printf("ERROR: LoginHandler: Error sending login request - %v\n", err)
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			log.Printf("LoginHandler: Invalid credentials for user %s - status code %d\n", username, resp.StatusCode)
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}

		authKey, _ := ioutil.ReadAll(resp.Body)
		var tokenResponse TokenResponse


		erro := json.Unmarshal([]byte(string(authKey)), &tokenResponse)
		if erro != nil {
			log.Fatal("ERROR: Error parsing JSON:", erro)
		}

		log.Printf("LoginHandler: Successfully authenticated user %s\n", username)

		// Set cookies for auth_key and username
		http.SetCookie(w, &http.Cookie{
			Name:     "auth_key",
			Value:    string(tokenResponse.AuthToken),
			Path:     "/",
			HttpOnly: true, // Security enhancement
			Expires:  time.Now().Add(24 * time.Hour), // Cookie expiration
		})
		http.SetCookie(w, &http.Cookie{
			Name:     "username",
			Value:    username,
			Path:     "/",
			HttpOnly: true, // Security enhancement
			Expires:  time.Now().Add(24 * time.Hour), // Cookie expiration
		})

		http.Redirect(w, r, "/userhome", http.StatusSeeOther)
		return
	}
	log.Println("LoginHandler: Serving login page")
	tmpl.ExecuteTemplate(w, "login.html", nil)
}

// UserHomeHandler serves the user home page with user details
func UserHomeHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("auth_key")
	if err != nil {
		log.Println("UserHomeHandler: Auth key cookie not found or expired")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	usernameCookie, err := r.Cookie("username")
	if err != nil {
		log.Println("UserHomeHandler: Username cookie not found or expired")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	log.Printf("UserHomeHandler: Fetching details for user %s\n", usernameCookie.Value)

	req, _ := http.NewRequest("GET", "http://userinfo:8083/userdetails", nil)
	req.Header.Set("Authorization", cookie.Value)
	req.Header.Set("username", usernameCookie.Value)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("ERROR: UserHomeHandler: Error sending request for user details - %v\n", err)
		http.Error(w, "Error fetching user details", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("ERROR: UserHomeHandler: Unexpected status code %d while fetching user details\n", resp.StatusCode)
		http.Error(w, "Error fetching user details", http.StatusInternalServerError)
		return
	}

	var userDetails UserDetails
	if err := json.NewDecoder(resp.Body).Decode(&userDetails); err != nil {
		log.Printf("ERROR: UserHomeHandler: Error parsing user details - %v\n", err)
		http.Error(w, "Error parsing user data", http.StatusInternalServerError)
		return
	}

	log.Printf("UserHomeHandler: Successfully fetched details for user %s\n", usernameCookie.Value)
	tmpl.ExecuteTemplate(w, "userhome.html", userDetails)
}

func SignUpHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		// Clear existing cookies
		http.SetCookie(w, &http.Cookie{
			Name:     "auth_key",
			Value:    "",
			Path:     "/",
			HttpOnly: true,
			Expires:  time.Now().Add(-24 * time.Hour), // Set to a past date to delete
		})
		http.SetCookie(w, &http.Cookie{
			Name:     "username",
			Value:    "",
			Path:     "/",
			HttpOnly: true,
			Expires:  time.Now().Add(-24 * time.Hour), // Set to a past date to delete
		})

		username := r.FormValue("username")
		_ = r.FormValue("password")
		email := r.FormValue("email")
		age := r.FormValue("age")

		//user := User{Username: username, Password: password}
		userDetails := UserDetails{Name: username, Email: email, Age: age}

		log.Printf("SignUpHandler: Registering new user %s\n", username)

		// First, send user details to register
		userJson, _ := json.Marshal(userDetails)
		resp, err := http.Post("http://userinfo:8083/useradd", "application/json", strings.NewReader(string(userJson)))
		if err != nil {
			log.Printf("ERROR: SignUpHandler: Error sending signup request - %v\n", err)
			http.Error(w, "Error signing up", http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		// Handle different response status codes
		if resp.StatusCode == http.StatusOK {
			log.Printf("SignUpHandler: Successfully added user %s\n", username)
			http.Redirect(w, r, "/login", http.StatusSeeOther)
		} else if resp.StatusCode == http.StatusConflict {
			// User already exists
			log.Printf("ERROR: SignUpHandler: User already exists - %s\n", username)
			http.Error(w, "User already exists", http.StatusConflict)
		} else {
			log.Printf("ERROR: SignUpHandler: Error response while signing up user %s - status code %d\n", username, resp.StatusCode)
			http.Error(w, "Error signing up", http.StatusInternalServerError)
		}

		return
	}

	log.Println("SignUpHandler: Serving signup page")
	// Serve the signup page template
	tmpl.ExecuteTemplate(w, "signup.html", nil)
}

func main() {
	defer logFile.Close() // Ensure log file is closed when main function exits
	http.Handle("/styles.css", http.FileServer(http.Dir(".")))
	http.HandleFunc("/", HomeHandler)
	http.HandleFunc("/login", LoginHandler)
	http.HandleFunc("/userhome", UserHomeHandler)
	http.HandleFunc("/signup", SignUpHandler)

	log.Println("Server started at http://webserver:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server failed to start - %v\n", err)
	}
}
