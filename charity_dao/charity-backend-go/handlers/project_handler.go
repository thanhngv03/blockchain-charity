package handlers

// Hiển thị dữ liệu lên Dashboard
import (
	"encoding/json"
	"net/http"
	"strconv"

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

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Lấy tham số từ URL (nếu có)
	walletParam := r.URL.Query().Get("wallet")
	statusParam := r.URL.Query().Get("status")

	// Xây dựng câu lệnh SQL động
	query := `
        SELECT id, title, description, target_amount, status, created_at 
        FROM projects 
        WHERE 1=1
    `
	var args []interface{}
	argId := 1

	if walletParam != "" {
		query += ` AND LOWER(creator_wallet) = LOWER($` + strconv.Itoa(argId) + `)`
		args = append(args, walletParam)
		argId++
	}

	if statusParam != "" {
		query += ` AND status = $` + strconv.Itoa(argId)
		args = append(args, statusParam)
		argId++
	}

	query += ` ORDER BY created_at DESC`

	rows, err := utils.DB.Query(query, args...)
	if err != nil {
		http.Error(w, "Database error: "+err.Error(), 500)
		return
	}
	defer rows.Close()

	var projects []Project
	for rows.Next() {
		var p Project
		var statusInt int

		if err := rows.Scan(&p.ID, &p.Title, &p.Description, &p.TargetAmount, &statusInt, &p.CreatedAt); err != nil {
			http.Error(w, "Scan error: "+err.Error(), 500)
			return
		}

		// QUY ƯỚC TRẠNG THÁI MỚI:
		// 0 = Chờ duyệt (pending)
		// 1 = Đang kêu gọi (calling) - Đã được Admin duyệt
		// 2 = Đã giải ngân (completed)
		switch statusInt {
		case 0:
			p.Status = "pending" // Đổi mặc định khi mới tạo thành pending
		case 1:
			p.Status = "calling"
		case 2:
			p.Status = "completed"
		default:
			p.Status = "pending"
		}

		p.CollectedAmount = 0
		projects = append(projects, p)
	}

	json.NewEncoder(w).Encode(projects)
}
