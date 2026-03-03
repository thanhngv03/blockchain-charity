package handlers

// Hiển thị dữ liệu lên Dashboard
import (
	"encoding/json"
	"net/http"

	"github.com/thanhngv03/decentralized-charity-fund/charity-backend-go/utils"
)

type Project struct {
	ID              string  `json:"id"`
	Title           string  `json:"title"`
	Description     string  `json:"description"`
	TargetAmount    float64 `json:"target_amount_wei"` // Map về tên FE đang dùng
	CollectedAmount float64 `json:"collected_amount_wei"`
	Status          string  `json:"status"` // Chúng ta sẽ convert số sang chữ ở đây
	CreatedAt       string  `json:"created_at"`
}

func GetProjectsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Query đúng các cột có trong image_6dc97f.png
	rows, err := utils.DB.Query(`
        SELECT id, title, description, target_amount, status, created_at 
        FROM projects 
        ORDER BY created_at DESC
    `)
	if err != nil {
		http.Error(w, "Database error: "+err.Error(), 500)
		return
	}
	defer rows.Close()

	var projects []Project

	for rows.Next() {
		var p Project
		var statusInt int

		// Scan khớp với các cột trong DB (id là uuid, status là int4)
		if err := rows.Scan(&p.ID, &p.Title, &p.Description, &p.TargetAmount, &statusInt, &p.CreatedAt); err != nil {
			http.Error(w, "Scan error: "+err.Error(), 500)
			return
		}

		switch statusInt {
		case 0:
			p.Status = "calling"
		case 1:
			p.Status = "pending"
		case 2:
			p.Status = "completed"
		default:
			p.Status = "calling"
		}
		// Giả định collected_amount = 0 nếu bạn chưa làm logic tính toán từ contract
		p.CollectedAmount = 0
		projects = append(projects, p)
	}

	json.NewEncoder(w).Encode(projects)
}
