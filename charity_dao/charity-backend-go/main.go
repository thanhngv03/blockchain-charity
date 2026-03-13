package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rs/cors"
	"github.com/thanhngv03/decentralized-charity-fund/charity-backend-go/handlers"
	"github.com/thanhngv03/decentralized-charity-fund/charity-backend-go/services"
	"github.com/thanhngv03/decentralized-charity-fund/charity-backend-go/utils"
)

// ===== Struct đọc JSON deploy =====
type DeploymentInfo struct {
	CharityVault string `json:"CharityVault"`
}

func main() {

	// 1. Connect PostgreSQL
	utils.ConnectDB()

	// 2. Load contract address
	data, err := os.ReadFile("../deployments/deployed-address.json")
	if err != nil {
		log.Fatal("Cannot read deploy info:", err)
	}

	var deployment DeploymentInfo
	if err := json.Unmarshal(data, &deployment); err != nil {
		log.Fatal("Invalid deploy JSON:", err)
	}

	expectedContract := common.HexToAddress(deployment.CharityVault)

	// 3. Connect Ethereum client
	client, err := ethclient.Dial("http://127.0.0.1:8545")
	if err != nil {
		log.Fatal("Cannot connect eth node:", err)
	}

	// 4. Health check
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// 5. Donate verify endpoint
	http.HandleFunc("/api/donations", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var body struct {
			Wallet string `json:"wallet_address"`
			TxHash string `json:"tx_hash"`
		}

		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "Invalid JSON body", http.StatusBadRequest)
			return
		}

		donor, amount, err := services.VerifyDonateTx(
			client,
			body.TxHash,
			expectedContract,
		)

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Save DB
		_, err = utils.DB.ExecContext(
			context.Background(),
			`INSERT INTO donations (tx_hash, wallet_address, amount_wei)
			VALUES ($1, $2, $3)
			ON CONFLICT (tx_hash) DO NOTHING`,
			body.TxHash,
			donor.Hex(),
			amount.String(),
		)

		if err != nil {
			http.Error(w, "Database error", 500)
			return
		}

		json.NewEncoder(w).Encode(map[string]string{
			"status": "success",
		})

	})

	// Project
	http.HandleFunc("/api/projects", handlers.GetProjectsHandler)
	http.HandleFunc("/api/admin/projects", handlers.CreateProjectHandler)
	http.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("uploads"))))
	http.HandleFunc("/api/admin/projects/delete", handlers.DeleteProjectHandler)
	http.HandleFunc("/api/admin/projects/update", handlers.UpdateProjectHandler)
	http.HandleFunc("/api/donations/history", handlers.GetDonationsHistoryHandler)
	http.HandleFunc("/api/donations/create", handlers.CreateDonationHandler)
	http.HandleFunc("/api/news", handlers.GetNewsFeed) // GET

	http.HandleFunc("/api/news/create", handlers.CreateNewsPost)  // POST (Nên bảo vệ bằng check Admin)
	http.HandleFunc("/api/news/like", handlers.ToggleLikeNews)    // POST
	http.HandleFunc("/api/news/comment", handlers.AddNewsComment) // POST

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000"}, // Cho phép React App
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Origin", "Content-Type", "Authorization", "wallet_address"},
		AllowCredentials: true,
		Debug:            true, // Bật cái này để bạn xem log CORS trong terminal Go
	})

	// Áp dụng cấu hình vào handler
	handler := c.Handler(http.DefaultServeMux)

	// 6. Start server
	fmt.Println("Backend running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}
