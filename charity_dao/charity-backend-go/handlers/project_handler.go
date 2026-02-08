package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/thanhngv03/decentralized-charity-fund/charity-backend-go/utils"
)

type Project struct {
	ID                 int    `json:"id"`
	Title              string `json:"title"`
	Description        string `json:"description"`
	TargetAmountWei    string `json:"target_amount_wei"`
	CollectedAmountWei string `json:"collected_amount_wei"`
	Status             string `json:"status"`
	CreatedAt          string `json:"created_at"`
	UpdatedAt          string `json:"updated_at"`
}

func GetProjectsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rows, err := utils.DB.Query(`
		SELECT id, title, description, target_amount_wei,
		       collected_amount_wei, status, created_at, updated_at
		FROM projects
		ORDER BY created_at DESC
	`)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer rows.Close()

	var projects []Project

	for rows.Next() {
		var p Project
		if err := rows.Scan(
			&p.ID,
			&p.Title,
			&p.Description,
			&p.TargetAmountWei,
			&p.CollectedAmountWei,
			&p.Status,
			&p.CreatedAt,
			&p.UpdatedAt,
		); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		projects = append(projects, p)
	}

	json.NewEncoder(w).Encode(projects)
}
