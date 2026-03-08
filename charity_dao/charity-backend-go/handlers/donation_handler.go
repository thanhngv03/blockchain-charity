package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/thanhngv03/decentralized-charity-fund/charity-backend-go/utils"
)


// Struct nhận/trả dữ liệu Quyên góp
type Donation struct {
	ID          string  `json:"id"`
	ProjectID   string  `json:"project_id"`
	ProjectName string  `json:"project_name"` // Lấy từ bảng projects
	DonorWallet string  `json:"donor_wallet"`
	DonorName   string  `json:"donor_name"`
	Message     string  `json:"message"`
	Amount      float64 `json:"amount_wei"`
	CreatedAt   string  `json:"created_at"`
	NetworkName string  `json:"network_name"` // Lấy từ bảng networks
}

// 1. HÀM LƯU QUYÊN GÓP (POST)
func CreateDonationHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var input Donation
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Dữ liệu không hợp lệ", http.StatusBadRequest)
		return
	}

	// Xử lý tên mặc định nếu user ẩn danh
	if input.DonorName == "" {
		input.DonorName = "Nhà hảo tâm"
	}

	query := `
		INSERT INTO donations (project_id, donor_wallet, donor_name, message, amount_wei)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := utils.DB.Exec(query, input.ProjectID, input.DonorWallet, input.DonorName, input.Message, input.Amount)
	if err != nil {
		http.Error(w, "Lỗi lưu database: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Ghi nhận quyên góp thành công!"})
}

// 2. HÀM LẤY LỊCH SỬ QUYÊN GÓP (GET)
func GetDonationsHistoryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// JOIN 3 bảng: donations, projects, networks để lấy đủ thông tin hiển thị
	query := `
		SELECT 
			d.id, d.project_id, p.title, d.donor_wallet, d.donor_name, d.message, d.amount_wei, d.created_at,
			COALESCE(n.network_name, 'Blockchain') as network_name
		FROM donations d
		JOIN projects p ON d.project_id = p.id
		LEFT JOIN networks n ON p.network_type_id = n.id
		ORDER BY d.created_at DESC
	`

	rows, err := utils.DB.Query(query)
	if err != nil {
		http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var donations []Donation
	for rows.Next() {
		var d Donation
		if err := rows.Scan(
			&d.ID, &d.ProjectID, &d.ProjectName, &d.DonorWallet, &d.DonorName,
			&d.Message, &d.Amount, &d.CreatedAt, &d.NetworkName,
		); err != nil {
			return
		}
		donations = append(donations, d)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(donations)
}
