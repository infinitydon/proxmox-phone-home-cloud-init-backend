package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"
)

type PhoneHomeRequest struct {
	ID        string `json:"id"`
	EventName string `json:"event_name"`
	Name      string `json:"name"`
	Result    string `json:"result"`
}

var db *sql.DB

func initDB() {
	var err error
	connStr := "user=go_proxmox password=go_proxmox2024 dbname=phone_home sslmode=disable"
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Error connecting to the database: %v", err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatalf("Error pinging the database: %v", err)
	}

	log.Println("Connected to the database")
}

func phoneHomeHandler(w http.ResponseWriter, r *http.Request) {
	var req PhoneHomeRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if req.EventName == "create" {
		req.Result = "success"
		_, err = db.Exec(
			"INSERT INTO instances (instance_id, event_name, name, result) VALUES ($1, $2, $3, $4)",
			req.ID, req.EventName, req.Name, req.Result,
		)
		if err == nil {
			log.Printf("Created instance: %+v\n", req)
		}
	} else if req.EventName == "delete" {
		_, err = db.Exec("DELETE FROM instances WHERE instance_id = $1", req.ID)
		if err == nil {
			log.Printf("Deleted instance with ID: %s\n", req.ID)
		}
	}

	if err != nil {
		log.Printf("Database operation failed: %v", err) // Log the error for troubleshooting
		http.Error(w, "Database operation failed", http.StatusInternalServerError)
		return
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

	var instance PhoneHomeRequest
	err := db.QueryRow("SELECT instance_id, event_name, name, result FROM instances WHERE instance_id = $1", id).Scan(
		&instance.ID, &instance.EventName, &instance.Name, &instance.Result,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Instance not found", http.StatusNotFound)
		} else {
			http.Error(w, "Database query failed", http.StatusInternalServerError)
		}
		return
	}

	log.Printf("Queried instance status: %+v\n", instance)

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(instance)
}

func listInstanceIDs(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT instance_id FROM instances")
	if err != nil {
		http.Error(w, "Database query failed", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		err := rows.Scan(&id)
		if err != nil {
			http.Error(w, "Error scanning row", http.StatusInternalServerError)
			return
		}
		ids = append(ids, id)
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ids)
}

func main() {
	initDB()
	defer db.Close()

	http.HandleFunc("/phone-home", phoneHomeHandler)
	http.HandleFunc("/instance-status", getInstanceStatus)
	http.HandleFunc("/list-ids", listInstanceIDs)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting server on :%s...", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
