package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

// Product represents the structure for a product item
type Product struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Price string `json:"price"`
	Link  string `json:"link"`
}

// Sample products data
var products = []Product{
	{ID: 1, Name: "Product 1", Price: "$100", Link: "/products/1"},
	{ID: 2, Name: "Product 2", Price: "$150", Link: "/products/2"},
	{ID: 3, Name: "Product 3", Price: "$200", Link: "/products/3"},
	{ID: 4, Name: "Product 4", Price: "$250", Link: "/products/4"},
	{ID: 5, Name: "Product 5", Price: "$300", Link: "/products/5"},
}

var (
	// Logger for writing to the console and log file
	logger *log.Logger
	logFile *os.File
)

// ProductsHandler handles requests to /products and responds with a list of all products
func ProductsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET method is allowed", http.StatusMethodNotAllowed)
		logger.Println("Invalid request method:", r.Method)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(products)
	if err != nil {
		http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
		logger.Println("Error encoding JSON:", err)
		return
	}

	logger.Println("Products list served")
}

// ProductDetailsHandler handles requests to /products/{id} and responds with details of a specific product
func ProductDetailsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET method is allowed", http.StatusMethodNotAllowed)
		logger.Println("Invalid request method:", r.Method)
		return
	}

	// Extract the product ID from the URL
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 3 {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		logger.Println("Invalid product ID in URL:", r.URL.Path)
		return
	}

	// Find the product based on the ID
	id := pathParts[2]
	for _, product := range products {
		if fmt.Sprintf("%d", product.ID) == id {
			err := json.NewEncoder(w).Encode(product)
			if err != nil {
				http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
				logger.Println("Error encoding JSON:", err)
				return
			}
			logger.Printf("Product details served for ID: %s\n", id)
			return
		}
	}

	// If product is not found, return a 404 error
	http.Error(w, "Product not found", http.StatusNotFound)
	logger.Printf("Product not found for ID: %s\n", id)
}

func main() {
	// Open log file
	var err error
	logFile, err = os.OpenFile("/logs/productlist.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		fmt.Println("Error opening log file:", err)
		os.Exit(1)
	}
	defer logFile.Close()

	// Set up logger
	logger = log.New(logFile, "", log.Ldate|log.Ltime|log.Lshortfile)

	http.HandleFunc("/products", ProductsHandler)
	http.HandleFunc("/products/", ProductDetailsHandler)

	fmt.Println("API server running on http://productlist:8081")
	logger.Println("API server running on http://productlist:8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}
