package main

import (
	"container/list"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
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

type LRUCache struct {
	mu       sync.Mutex
	capacity int
	cache    map[int]*list.Element
	lruList  *list.List
}

func NewLRUCache(capacity int) *LRUCache {
	return &LRUCache{
		capacity: capacity,
		cache:    make(map[int]*list.Element),
		lruList:  list.New(),
	}
}

func (c *LRUCache) Get(id int) (*ProductCompare, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if elem, ok := c.cache[id]; ok {
		c.lruList.MoveToFront(elem)
		return elem.Value.(*ProductCompare), true
	}
	return nil, false
}

func (c *LRUCache) Put(id int, value *ProductCompare) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if elem, ok := c.cache[id]; ok {
		c.lruList.MoveToFront(elem)
		elem.Value = value
		return
	}
	if c.lruList.Len() >= c.capacity {
		delete(c.cache, c.removeOldest())
	}
	elem := c.lruList.PushFront(value)
	c.cache[id] = elem
}

func (c *LRUCache) removeOldest() int {
	elem := c.lruList.Back()
	if elem != nil {
		c.lruList.Remove(elem)
		return elem.Value.(*Product).ID
	}
	return 0
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
	// Parse query parameters
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

	// Fetch product information for both products from the cache or API
	cache := NewLRUCache(10)
	product1, ok1 := cache.Get(productID1)
	if !ok1 {
		product1, err = FetchProductByID(productID1)
		if err != nil {
			http.Error(w, "Failed to fetch product1", http.StatusInternalServerError)
			return
		}
		cache.Put(productID1, product1)
	}

	product2, ok2 := cache.Get(productID2)
	if !ok2 {
		product2, err = FetchProductByID(productID2)
		if err != nil {
			http.Error(w, "Failed to fetch product2", http.StatusInternalServerError)
			return
		}
		cache.Put(productID2, product2)
	}

	// Compare prices of the two products
	var priceWinner string
	if product1.Price > product2.Price {
		priceWinner = strconv.Itoa(productID1)
	} else if product1.Price < product2.Price {
		priceWinner = strconv.Itoa(productID2)
	}

	// Create response JSON
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

	// Write response JSON
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
