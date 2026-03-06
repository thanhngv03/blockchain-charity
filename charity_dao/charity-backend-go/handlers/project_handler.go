package handlers

// Hiển thị dữ liệu lên Dashboard
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
	CollectedAmount    float64 `json:"collected_amount_wei"`
	Status             string  `json:"status"`
	CreatedAt          string  `json:"created_at"`
	Image              string  `json:"image"`
	IdFiles            string  `json:"id_files"`
	BeneficiaryName    string  `json:"beneficiary_name"`
	BeneficiaryContact string  `json:"beneficiary_contact"`
	Province           string  `json:"province"`
	District           string  `json:"district"`
	Address            string  `json:"address"`
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
        SELECT id, title, description, target_amount, status, created_at, image
		, id_files, beneficiary_name, beneficiary_contact, province, district, address 
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

		var image, id_files, beneficiary_name, beneficiary_contact, province, district, address sql.NullString

		if err := rows.Scan(
			&p.ID, &p.Title, &p.Description, &p.TargetAmount, &statusInt, &p.CreatedAt,
			&image, &id_files, &beneficiary_name, &beneficiary_contact, &province, &district, &address,
		); err != nil {
			http.Error(w, "Scan error: "+err.Error(), 500)
			return
		}

		p.Image = image.String
		p.IdFiles = id_files.String
		p.BeneficiaryName = beneficiary_name.String
		p.BeneficiaryContact = beneficiary_contact.String
		p.Province = province.String
		p.District = district.String
		p.Address = address.String
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
