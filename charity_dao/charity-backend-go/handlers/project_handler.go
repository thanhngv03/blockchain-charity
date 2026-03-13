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
	EndDate            string  `json:"end_date"`
}

func GetProjectsHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")

    if r.Method != http.MethodGet {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    // 1. TỰ ĐỘNG CẬP NHẬT TRẠNG THÁI HẾT HẠN TRƯỚC KHI LẤY DỮ LIỆU
    updateExpiryQuery := `
        UPDATE projects 
        SET status = 3 
        WHERE status = 1 
          AND end_date IS NOT NULL 
          AND end_date < CURRENT_TIMESTAMP
    `
    utils.DB.Exec(updateExpiryQuery)

    walletParam := r.URL.Query().Get("wallet")
    statusParam := r.URL.Query().Get("status")

    // 2. CÂU LỆNH SELECT (Tổng cộng 18 cột - Đã đánh số để bạn dễ kiểm tra)
    query := `
        SELECT 
            p.id, p.title, p.description, p.target_amount, p.status, -- 1,2,3,4,5
            p.created_at, p.image, p.id_files, p.beneficiary_name, p.beneficiary_contact, -- 6,7,8,9,10
            p.province, p.district, p.address, p.receiver_wallet, -- 11,12,13,14
            n.network_name, -- 15
            (SELECT COUNT(*) FROM donations d WHERE d.project_id = p.id) as donation_count, -- 16
            (SELECT COALESCE(SUM(amount_wei), 0) FROM donations d WHERE d.project_id = p.id) as raised_amount, -- 17
            p.end_date -- 18 (CỘT MỚI THÊM)
        FROM projects p
        LEFT JOIN networks n ON p.network_type_id = n.id
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
        // Dùng NullString/NullTime để tránh lỗi khi dữ liệu trong DB bị trống (NULL)
        var image, idFiles, bName, bContact, prov, dist, addr, rWallet, nName sql.NullString
        var endDate sql.NullTime // Cần dùng NullTime cho cột TIMESTAMP

        // 3. HÀM SCAN (Phải đủ 18 biến và đúng thứ tự với SELECT ở trên)
        err := rows.Scan(
            &p.ID, &p.Title, &p.Description, &p.TargetAmount, &statusInt, // 1,2,3,4,5
            &p.CreatedAt, &image, &idFiles, &bName, &bContact, // 6,7,8,9,10
            &prov, &dist, &addr, &rWallet, // 11,12,13,14
            &nName, // 15
            &p.DonationCount, // 16
            &p.CollectedAmount, // 17
            &endDate, // 18 (BIẾN MỚI THÊM)
        )
        
        if err != nil {
            // Nếu vẫn lỗi, nó sẽ in ra lỗi cụ thể ở đây (ví dụ: expected 18 arguments, got 17)
            http.Error(w, "Scan error: "+err.Error(), 500)
            return
        }

        // Gán lại dữ liệu từ Null types sang Project struct
        p.Image = image.String
        p.IdFiles = idFiles.String
        p.BeneficiaryName = bName.String
        p.BeneficiaryContact = bContact.String
        p.Province = prov.String
        p.District = dist.String
        p.Address = addr.String
        p.ReceiverWallet = rWallet.String
        p.NetworkName = nName.String
        
        if endDate.Valid {
            p.EndDate = endDate.Time.Format("2006-01-02 15:04:05")
        } else {
            p.EndDate = "" // Nếu chưa có hạn thì để trống
        }

        // Logic hiển thị trạng thái chữ
        switch statusInt {
        case 0: p.Status = "pending"
        case 1: p.Status = "calling"
        case 2: p.Status = "paused"
        case 3: p.Status = "completed"
        default: p.Status = "pending"
        }

        projects = append(projects, p)
    }

    json.NewEncoder(w).Encode(projects)
}