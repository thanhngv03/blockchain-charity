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
	ProjectName string  `json:"project_name"`
	DonorWallet string  `json:"donor_wallet"`
	DonorName   string  `json:"donor_name"`
	Message     string  `json:"message"`
	Amount      float64 `json:"amount_wei"`
	TxHash      string  `json:"tx_hash"` // BỔ SUNG: Trường nhận mã giao dịch từ React
	CreatedAt   string  `json:"created_at"`
	NetworkName string  `json:"network_name"`
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

	// --- BƯỚC 1: LƯU LỊCH SỬ QUYÊN GÓP CÙNG TX_HASH ---
	queryInsert := `
        INSERT INTO donations (project_id, donor_wallet, donor_name, message, amount_wei, tx_hash)
        VALUES ($1, $2, $3, $4, $5, $6)
    `
	_, err := utils.DB.Exec(queryInsert, input.ProjectID, input.DonorWallet, input.DonorName, input.Message, input.Amount, input.TxHash)
	if err != nil {
		http.Error(w, "Lỗi lưu database lịch sử: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// --- BƯỚC 2: CẬP NHẬT TỔNG TIỀN VÀ LƯỢT QUYÊN GÓP CHO DỰ ÁN ---
	// Dùng COALESCE để nếu raised_amount hoặc donation_count đang là NULL thì gán bằng 0 trước khi cộng
	queryUpdate := `
        UPDATE projects 
        SET raised_amount = COALESCE(raised_amount, 0) + $1,
            donation_count = COALESCE(donation_count, 0) + 1
        WHERE id = $2
    `
	_, err = utils.DB.Exec(queryUpdate, input.Amount, input.ProjectID)
	if err != nil {
		http.Error(w, "Lỗi cập nhật số liệu dự án: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Ghi nhận quyên góp và cập nhật dự án thành công!"})
}

// 2. HÀM LẤY LỊCH SỬ QUYÊN GÓP (GET)
func GetDonationsHistoryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Dùng COALESCE(d.tx_hash, '') để tránh lỗi Go Crash khi đọc các dữ liệu cũ bị NULL ở cột tx_hash
	query := `
        SELECT 
            d.id, d.project_id, p.title, d.donor_wallet, d.donor_name, d.message, d.amount_wei, 
            COALESCE(d.tx_hash, '') as tx_hash, d.created_at,
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
		// Cần truyền BIẾN vào ĐÚNG THỨ TỰ của câu SELECT ở trên
		if err := rows.Scan(
			&d.ID, &d.ProjectID, &d.ProjectName, &d.DonorWallet, &d.DonorName,
			&d.Message, &d.Amount, &d.TxHash, &d.CreatedAt, &d.NetworkName,
		); err != nil {
			http.Error(w, "Lỗi map dữ liệu: "+err.Error(), http.StatusInternalServerError)
			return
		}
		donations = append(donations, d)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(donations)
}
