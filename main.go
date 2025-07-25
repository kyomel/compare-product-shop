package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

type Product struct {
	ID          int     `json:"id"`
	Title       string  `json:"title"`
	Price       float64 `json:"price"`
	Description string  `json:"description"`
	Category    string  `json:"category"`
	Image       string  `json:"image"`
}

type ProductCompare struct {
	ID    int     `json:"id"`
	Price float64 `json:"price"`
}

func FetchProductByID(productID int) (*ProductCompare, error) {
	url := fmt.Sprintf("https://fakestoreapi.com/products/%d", productID)
	fmt.Println(url)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var product ProductCompare
	if err := json.NewDecoder(resp.Body).Decode(&product); err != nil {
		return nil, err
	}

	return &ProductCompare{ID: product.ID, Price: product.Price}, nil
}

func CompareProductsHandler(w http.ResponseWriter, r *http.Request) {
	productID1, err := strconv.Atoi(r.URL.Query().Get("productID1"))
	if err != nil {
		http.Error(w, "Invalid productID1", http.StatusBadRequest)
		return
	}

	productID2, err := strconv.Atoi(r.URL.Query().Get("productID2"))
	if err != nil {
		http.Error(w, "Invalid productID2", http.StatusBadRequest)
		return
	}

	product1, err := FetchProductByID(productID1)
	fmt.Println(product1)

	if err != nil {
		http.Error(w, "Failed to fetch product 1", http.StatusInternalServerError)
		return
	}

	product2, err := FetchProductByID(productID2)
	fmt.Println(product2)

	if err != nil {
		http.Error(w, "Failed to fetch product 2", http.StatusInternalServerError)
		return
	}

	priceWinner := ""
	if product1.Price < product2.Price {
		priceWinner = strconv.Itoa(product1.ID)
	} else if product1.Price > product2.Price {
		priceWinner = strconv.Itoa(product2.ID)
	}

	response := struct {
		Price struct {
			ProductOne float64 `json:"productOne"`
			ProductTwo float64 `json:"productTwo"`
			Winner     string  `json:"winner"`
		} `json:"price"`
	}{
		Price: struct {
			ProductOne float64 `json:"productOne"`
			ProductTwo float64 `json:"productTwo"`
			Winner     string  `json:"winner"`
		}{
			ProductOne: product1.Price,
			ProductTwo: product2.Price,
			Winner:     priceWinner,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func FetchProducts() ([]Product, error) {
	resp, err := http.Get("https://fakestoreapi.com/products")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var products []Product
	if err := json.NewDecoder(resp.Body).Decode(&products); err != nil {
		return nil, err
	}

	return products, nil
}

func ListProductsHandler(w http.ResponseWriter, r *http.Request) {
	products, err := FetchProducts()
	if err != nil {
		http.Error(w, "Failed to fetch products", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(products); err != nil {
		http.Error(w, "Failed to encode products", http.StatusInternalServerError)
		return
	}
}

func main() {
	http.HandleFunc("/products", ListProductsHandler)
	http.HandleFunc("/compare", CompareProductsHandler)

	log.Printf("Server started on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
