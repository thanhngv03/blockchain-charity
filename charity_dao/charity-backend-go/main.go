package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/thanhngv03/decentralized-charity-fund/charity-backend-go/db"
	"github.com/thanhngv03/decentralized-charity-fund/charity-backend-go/handlers"
)

func main() {

	// 1. Connect PostgreSQL
	if err := db.ConnectPostgres(); err != nil {
		log.Fatal("PostgreSQL connection failed:", err)
	}
	log.Println("PostgreSQL connected")

	// 2. Health check
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	http.HandleFunc("/api/donations", handlers.CreateDonation)
	http.HandleFunc("/api/stats", handlers.StatsHandler)

	// 3. Start server
	fmt.Println("Backend running at http://localhost:3000")
	log.Fatal(http.ListenAndServe(":3000", nil))

}
