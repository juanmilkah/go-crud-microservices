package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
)

type Item struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

var (
	items  = make(map[string]Item)
	mutex  = &sync.Mutex{}
	nextID = 1
)

func main() {
	http.HandleFunc("/items", itemsHandler)
	http.HandleFunc("/items/", itemHandler)

	fmt.Println("Server is running on port 8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

func itemsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getItems(w)
	case http.MethodPost:
		createItem(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func itemHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/items/"):]

	switch r.Method {
	case http.MethodGet:
		getItem(w, id)
	case http.MethodPut:
		updateItem(w, r, id)
	case http.MethodDelete:
		deleteItem(w, id)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func getItems(w http.ResponseWriter) {
	mutex.Lock()
	defer mutex.Unlock()

	var itemList []Item
	for _, item := range items {
		itemList = append(itemList, item)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(itemList)
}

func createItem(w http.ResponseWriter, r *http.Request) {
	var newItem Item
	if err := json.NewDecoder(r.Body).Decode(&newItem); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	mutex.Lock()
	newItem.ID = fmt.Sprintf("%d", nextID)
	nextID++
	items[newItem.ID] = newItem
	mutex.Unlock()

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newItem)
}

func getItem(w http.ResponseWriter, id string) {
	mutex.Lock()
	item, exists := items[id]
	mutex.Unlock()

	if !exists {
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(item)
}

func updateItem(w http.ResponseWriter, r *http.Request, id string) {
	var updatedItem Item
	if err := json.NewDecoder(r.Body).Decode(&updatedItem); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	mutex.Lock()
	defer mutex.Unlock()

	if _, exists := items[id]; !exists {
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}

	updatedItem.ID = id
	items[id] = updatedItem
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedItem)
}

func deleteItem(w http.ResponseWriter, id string) {
	mutex.Lock()
	defer mutex.Unlock()

	if _, exists := items[id]; !exists {
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}

	delete(items, id)
	w.WriteHeader(http.StatusNoContent)
}
