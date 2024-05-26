package main

import (
        "encoding/json"
        "log"
        "net/http"
        "sync"
)

type PhoneHomeRequest struct {
        ID           string `json:"id"`
        EventName    string `json:"event_name"`
        LocalHost    string `json:"local_hostname,omitempty"`
        PublicIPv4   string `json:"public_ipv4,omitempty"`
        PrivateIPv4  string `json:"private_ipv4,omitempty"`
        Result       string `json:"result,omitempty"`
        Name         string `json:"name,omitempty"`
        Type         string `json:"type,omitempty"`
        Region       string `json:"region,omitempty"`
        KeyPair      string `json:"keypair,omitempty"`
}

var (
        instances = make(map[string]PhoneHomeRequest)
        mu        sync.Mutex
)

func phoneHomeHandler(w http.ResponseWriter, r *http.Request) {
        var req PhoneHomeRequest
        err := json.NewDecoder(r.Body).Decode(&req)
        if err != nil {
                http.Error(w, "Invalid request payload", http.StatusBadRequest)
                return
        }

        mu.Lock()
        defer mu.Unlock()

        if req.EventName == "create" {
                req.Result = "success"
                instances[req.ID] = req
                log.Printf("Created instance: %v\n", req)
        } else if req.EventName == "delete" {
                delete(instances, req.ID)
                log.Printf("Deleted instance with ID: %s\n", req.ID)
        }

        w.WriteHeader(http.StatusOK)
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(req)
}

func getInstanceStatus(w http.ResponseWriter, r *http.Request) {
        id := r.URL.Query().Get("id")
        if id == "" {
                http.Error(w, "Missing id parameter", http.StatusBadRequest)
                return
        }

        mu.Lock()
        defer mu.Unlock()
        inst, exists := instances[id]
        if !exists {
                http.Error(w, "Instance not found", http.StatusNotFound)
                return
        }

        log.Printf("Queried instance status: %v\n", inst)
        w.WriteHeader(http.StatusOK)
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(inst)
}

func listInstanceIDs(w http.ResponseWriter, r *http.Request) {
        mu.Lock()
        defer mu.Unlock()
        ids := make([]string, 0, len(instances))
        for id := range instances {
                ids = append(ids, id)
        }

        log.Printf("List of instance IDs: %v\n", ids)
        w.WriteHeader(http.StatusOK)
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(ids)
}

func main() {
        http.HandleFunc("/phone-home", phoneHomeHandler)
        http.HandleFunc("/instance-status", getInstanceStatus)
        http.HandleFunc("/list-ids", listInstanceIDs)

        log.Println("Starting server on :8080...")
        if err := http.ListenAndServe(":8080", nil); err != nil {
                log.Fatalf("Server failed to start: %v", err)
        }
}
