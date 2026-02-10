package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/thanhngv03/decentralized-charity-fund/charity-backend-go/utils"
)

func CreateProjectHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", 405)
		return
	}

	var p struct {
		Title           string `json:"title"`
		Description     string `json:"description"`
		Category        string `json:"category"`
		TargetAmountWei string `json:"target_amount_wei"`
		WalletAddress   string `json:"wallet_address"`
		OwnerName       string `json:"owner_name"`
		OwnerEmail      string `json:"owner_email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, "Invalid body", 400)
		return
	}

	_, err := utils.DB.Exec(`
		INSERT INTO projects
		(title, description, category, target_amount_wei, wallet_address,
		 owner_name, owner_email, status, approved)
		VALUES ($1,$2,$3,$4,$5,$6,$7,'draft',false)
	`,
		p.Title, p.Description, p.Category,
		p.TargetAmountWei, p.WalletAddress,
		p.OwnerName, p.OwnerEmail,
	)

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"message": "Project created",
	})
}

func UpdateProjectHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Lấy id từ query ?id=1
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Missing project id", http.StatusBadRequest)
		return
	}

	var body struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Status      string `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	_, err := utils.DB.Exec(`
		UPDATE projects
		SET title=$1,
		    description=$2,
		    status=$3,
		    updated_at=NOW()
		WHERE id=$4
	`,
		body.Title,
		body.Description,
		body.Status,
		id,
	)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"message": "Project updated successfully",
	})
}
