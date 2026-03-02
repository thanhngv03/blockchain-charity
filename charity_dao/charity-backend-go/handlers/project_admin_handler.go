package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/thanhngv03/decentralized-charity-fund/charity-backend-go/utils"
)

// DTO (Data Transfer Object) để nhận request từ Client
type CreateProjectRequest struct {
	Title         string `json:"title"`
	CategoryID    int    `json:"category_id"` // Thay đổi thành int (FK)
	Description   string `json:"description"`
	CreatorWallet string `json:"creator_wallet"` // Địa chỉ ví đăng nhập

	// Beneficiary Info
	BeneficiaryName    string                 `json:"beneficiary_name"`
	BeneficiaryContact string                 `json:"beneficiary_contact"`
	IDFiles            map[string]interface{} `json:"id_files"` // Nhận dạng JSON Object

	// Address
	Address  string `json:"address"`
	District string `json:"district"`
	Province string `json:"province"`

	// Blockchain & Finance
	TargetAmount      float64 `json:"target_amount"`   // Đổi thành số thực (Decimal trong DB)
	NetworkTypeID     int     `json:"network_type_id"` // FK
	ReceiverWallet    string  `json:"receiver_wallet"`
	PayoutConditionID int     `json:"payout_condition_id"` // FK

	// Media
	Links map[string]interface{} `json:"links,omitempty"` // Link ảnh mô tả
}

func CreateProjectHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var body CreateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid body: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Chuyển đổi JSONB
	idFilesJSON, _ := json.Marshal(body.IDFiles)
	linksJSON, _ := json.Marshal(body.Links)

	/* Deploy contract
	targetAmountStr := strconv.FormatFloat(body.TargetAmount, 'f', -1, 64)
	contractAddress, err := services.DeployProjectContract(targetAmountStr)
	if err != nil {
		http.Error(w, "Failed to deploy contract: "+err.Error(), http.StatusInternalServerError)
		return
	}
	*/

	tempContractAddress := "0x0000000000000000000000000000000000000000"

	// ===== SQL ĐÃ LOẠI BỎ is_private_docs VÀ KHỚP 17 THAM SỐ =====
	query := `
        INSERT INTO projects (
            title, category_id, description, creator_wallet,
            beneficiary_name, beneficiary_contact, id_files,
            address, district, province,
            target_amount, network_type_id, receiver_wallet, payout_condition_id, 
            contract_address, status, links
        ) 
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17) 
        RETURNING id`

	var newProjectID string
	// Truyền chính xác 17 biến tương ứng với $1 -> $17
	var err error
	err = utils.DB.QueryRow(query,
		body.Title,              // $1
		body.CategoryID,         // $2
		body.Description,        // $3
		body.CreatorWallet,      // $4
		body.BeneficiaryName,    // $5
		body.BeneficiaryContact, // $6
		string(idFilesJSON),     // $7
		body.Address,            // $8
		body.District,           // $9
		body.Province,           // $10
		body.TargetAmount,       // $11
		body.NetworkTypeID,      // $12
		body.ReceiverWallet,     // $13
		body.PayoutConditionID,  // $14
		tempContractAddress,     // Thay contractAddress.Hex() bằng biến tạm này
		0,
		string(linksJSON),
	).Scan(&newProjectID)

	if err != nil {
		// Trả về lỗi chi tiết để Frontend hiển thị
		http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"message":    "Project created successfully",
		"project_id": newProjectID,
	})
}

// -------------------------------------------------------------

type UpdateProjectRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      int    `json:"status"` // Sửa thành INT cho khớp Database mới
}

func UpdateProjectHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Lấy id từ query ?id=UUID
	id := r.URL.Query().Get("id")
	if strings.TrimSpace(id) == "" {
		http.Error(w, "Missing project id", http.StatusBadRequest)
		return
	}

	var body UpdateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid JSON body: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Câu lệnh UPDATE trỏ đúng các trường trong Database
	query := `
		UPDATE projects 
		SET title = $1, 
			description = $2, 
			status = $3, 
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $4
	`
	res, err := utils.DB.Exec(query, body.Title, body.Description, body.Status, id)

	if err != nil {
		http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"message": "Project updated successfully",
	})
}
