package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/thanhngv03/decentralized-charity-fund/charity-backend-go/utils"
)

func GetProjects(w http.ResponseWriter, r *http.Request) {
	rows, err := utils.DB.Query(`
		SELECT id, title, description, target_amount_wei, collected_amount_wei, status
		FROM projects
		WHERE status = 'active'
	`)
	if err != nil {
		http.Error(w, "DB error", 500)
		return
	}
	defer rows.Close()

	var projects []map[string]interface{}

	for rows.Next() {
		var id int
		var title, desc, status string
		var target, collected string

		rows.Scan(&id, &title, &desc, &target, &collected, &status)

		projects = append(projects, map[string]interface{}{
			"id":          id,
			"title":       title,
			"description": desc,
			"target":      target,
			"collected":   collected,
			"status":      status,
		})
	}

	json.NewEncoder(w).Encode(projects)
}
