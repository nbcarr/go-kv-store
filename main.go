package main

import (
	"fmt"
	"net/http"
	"sync"
	"encoding/json"
	"os"
	"time"
	"strconv"
)

type KVStore struct {
	data map[string]KeyValue
	mu sync.Mutex
	filepath string
}

type KeyValue struct {
	Value string
	ExpiresAt time.Time
}

func (kv *KVStore) save() error {
	data, err := json.Marshal(kv.data)
	if err != nil {
		fmt.Println("Error:", err)
		return err
	}
	return os.WriteFile(kv.filepath, data, 0644)
}

func (kv *KVStore) load() error {
    data, err := os.ReadFile(kv.filepath)
    if err != nil {
        return err
    }
    return json.Unmarshal(data, &kv.data)
}

func (kv *KVStore) get(w http.ResponseWriter, r *http.Request) {
	keys := r.URL.Query()["key"]
	if len(keys) == 0 {
		http.Error(w, "missing key parameter", http.StatusBadRequest)
		return
	}
 
	kv.mu.Lock()
	defer kv.mu.Unlock()
	var values []string
	for _, key := range keys {
		if val, exists := kv.data[key]; exists {
			if time.Now().After(val.ExpiresAt) {
				delete(kv.data, key)
				continue
			}
			values = append(values, fmt.Sprintf("%s: %s", key, val.Value))
		}
	}

	fmt.Fprintf(w, "%v", values)
 }

 func (kv *KVStore) put(w http.ResponseWriter, r *http.Request) {
    key := r.URL.Query().Get("key")
    value := r.URL.Query().Get("value")
    ttl := r.URL.Query().Get("ttl")
    defaultTTL := 86400 * 30 // 1 month
    var ttlSeconds int

    if key == "" || value == "" {
        http.Error(w, "missing key or value", http.StatusBadRequest)
        return
    }

    if ttl == "" {
        ttlSeconds = defaultTTL
        fmt.Printf("missing ttl, using default %d seconds\n", defaultTTL)
    } else {
        ttlSeconds, _ = strconv.Atoi(ttl)
    }

    kv.mu.Lock()
    defer kv.mu.Unlock()
    
    expiresAt := time.Now().Add(time.Duration(ttlSeconds) * time.Second)
    kv.data[key] = KeyValue{
        Value: value,
        ExpiresAt: expiresAt,
    }

    if err := kv.save(); err != nil {
        http.Error(w, "error saving data", http.StatusInternalServerError)
        return
    }
    
    w.WriteHeader(http.StatusCreated)
    fmt.Fprintf(w, "key %s added successfully", key)
}

func (kv *KVStore) delete(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	
	if key == "" {
		http.Error(w, "missing key", http.StatusBadRequest)
		return
	}
 
	kv.mu.Lock()
	defer kv.mu.Unlock()

	if _, exists := kv.data[key]; !exists {
		http.Error(w, "key not found", http.StatusNotFound)
		return
	}
	
	delete(kv.data, key)

	if err := kv.save(); err != nil {
        http.Error(w, "error saving data", http.StatusInternalServerError)
        return
    }
	
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "key %s deleted successfully", key)
 }

func homepage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "TODO: add usage\n")	
}

func (kv *KVStore) cleanExpired() {
    for k, v := range kv.data {
        if time.Now().After(v.ExpiresAt) {
            delete(kv.data, k)
        }
    }
}

func main() {
    kv := &KVStore{
        data: make(map[string]KeyValue),
		filepath: "store.json",
    }

	kv.cleanExpired()

	if err := kv.load(); err != nil {
		fmt.Printf("No existing data found: %v\n", err)
	}
	
	http.HandleFunc("/", homepage)
	http.HandleFunc("/GET", kv.get)
	http.HandleFunc("/PUT", kv.put)
	http.HandleFunc("/DELETE", kv.delete)

	fmt.Println("Server listening on port 8080...")
	http.ListenAndServe(":8080", nil)
}