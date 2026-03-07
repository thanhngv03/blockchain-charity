package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/thanhngv03/decentralized-charity-fund/charity-backend-go/utils"
)

type Project struct {
	ID                 string  `json:"id"`
	Title              string  `json:"title"`
	Description        string  `json:"description"`
	TargetAmount       float64 `json:"target_amount_wei"`
	CollectedAmount    float64 `json:"raised_amount"`  // Đổi tên cho khớp với frontend
	DonationCount      int     `json:"donation_count"` // Thêm trường đếm lượt quyên góp
	Status             string  `json:"status"`
	CreatedAt          string  `json:"created_at"`
	Image              string  `json:"image"`
	IdFiles            string  `json:"id_files"`
	BeneficiaryName    string  `json:"beneficiary_name"`
	BeneficiaryContact string  `json:"beneficiary_contact"`
	Province           string  `json:"province"`
	District           string  `json:"district"`
	Address            string  `json:"address"`
	ReceiverWallet     string  `json:"receiver_wallet"`
	NetworkTypeID      string  `json:"network_type_id"`
	NetworkName        string  `json:"network_name"`
}

func GetProjectsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	walletParam := r.URL.Query().Get("wallet")
	statusParam := r.URL.Query().Get("status")

	// CẬP NHẬT CÂU LỆNH SQL:
	// 1. JOIN với bảng networks (giả định bảng tên là networks)
	// 2. Dùng Subquery để tính COUNT và SUM từ bảng donations
	query := `
		SELECT 
			p.id, p.title, p.description, p.target_amount, p.status, p.created_at, p.image,
			p.id_files, p.beneficiary_name, p.beneficiary_contact, p.province, p.district, p.address,
			p.receiver_wallet, 
			n.network_name, -- Lấy tên mạng từ bảng networks
			(SELECT COUNT(*) FROM donations d WHERE d.project_id = p.id) as donation_count,
			(SELECT COALESCE(SUM(amount_wei), 0) FROM donations d WHERE d.project_id = p.id) as raised_amount
		FROM projects p
		LEFT JOIN networks n ON p.network_type_id = n.id -- Thực hiện JOIN
		WHERE 1=1
	`
	var args []interface{}
	argId := 1

	if walletParam != "" {
		query += ` AND LOWER(p.creator_wallet) = LOWER($` + strconv.Itoa(argId) + `)`
		args = append(args, walletParam)
		argId++
	}

	if statusParam != "" {
		query += ` AND p.status = $` + strconv.Itoa(argId)
		args = append(args, statusParam)
		argId++
	}

	query += ` ORDER BY p.created_at DESC`

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
		// Khai báo các biến NullString để xử lý dữ liệu trống
		var image, id_files, beneficiary_name, beneficiary_contact, province, district, address, receiver_wallet, network_name sql.NullString

		// THỨ TỰ TRONG SCAN PHẢI KHỚP 100% VỚI SELECT TRÊN
		err := rows.Scan(
			&p.ID, &p.Title, &p.Description, &p.TargetAmount, &statusInt, &p.CreatedAt, &image,
			&id_files, &beneficiary_name, &beneficiary_contact, &province, &district, &address,
			&receiver_wallet,
			&network_name,
			&p.DonationCount,
			&p.CollectedAmount,
		)
		if err != nil {
			http.Error(w, "Scan error: "+err.Error(), 500)
			return
		}

		// Gán lại giá trị từ NullString sang String
		p.Image = image.String
		p.IdFiles = id_files.String
		p.BeneficiaryName = beneficiary_name.String
		p.BeneficiaryContact = beneficiary_contact.String
		p.Province = province.String
		p.District = district.String
		p.Address = address.String
		p.ReceiverWallet = receiver_wallet.String
		p.NetworkName = network_name.String

		// Xử lý logic hiển thị trạng thái
		switch statusInt {
		case 0:
			p.Status = "pending"
		case 1:
			p.Status = "calling"
		case 2:
			p.Status = "paused" // Theo yêu cầu Tạm dừng ở tin nhắn trước
		case 3:
			p.Status = "completed"
		default:
			p.Status = "pending"
		}

		projects = append(projects, p)
	}

	json.NewEncoder(w).Encode(projects)
}
