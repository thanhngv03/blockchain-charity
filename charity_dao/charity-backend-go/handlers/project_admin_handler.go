package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/thanhngv03/decentralized-charity-fund/charity-backend-go/utils"
)

// DTO (Data Transfer Object) để nhận request từ Client
type CreateProjectRequest struct {
	Title         string `json:"title"`
	CategoryID    int    `json:"category_id"` // Thay đổi thành int (FK)
	Description   string `json:"description"`
	Image         string `json:"image"`          // Link ảnh mô tả
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

func saveUploadedFile(r *http.Request, formKey string) (string, error) {
	file, handler, err := r.FormFile(formKey)
	if err != nil {
		if err == http.ErrMissingFile {
			return "", nil // Không có file cũng không sao (nếu không bắt buộc)
		}
		return "", err
	}
	defer file.Close()

	// Tạo thư mục uploads nếu chưa có
	err = os.MkdirAll("uploads", os.ModePerm)
	if err != nil {
		return "", err
	}

	// Tạo tên file duy nhất để tránh trùng lặp
	fileName := fmt.Sprintf("%d_%s", time.Now().UnixNano(), handler.Filename)
	filePath := filepath.Join("uploads", fileName)

	// Tạo file trên server
	dst, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	// Copy nội dung vào file
	if _, err := io.Copy(dst, file); err != nil {
		return "", err
	}

	// Trả về đường dẫn tĩnh (để frontend có thể hiển thị)
	return "/uploads/" + fileName, nil
}

func CreateProjectHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 1. Phân tích multipart form (giới hạn 10MB)
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "Error parsing form data: "+err.Error(), http.StatusBadRequest)
		return
	}

	// 2. Lấy các trường text từ Form
	title := r.FormValue("title")
	categoryID, _ := strconv.Atoi(r.FormValue("category_id"))
	description := r.FormValue("description")
	creatorWallet := r.FormValue("creator_wallet")
	beneficiaryName := r.FormValue("beneficiary_name")
	beneficiaryContact := r.FormValue("beneficiary_contact")
	address := r.FormValue("address")
	district := r.FormValue("district")
	province := r.FormValue("province")
	targetAmount, _ := strconv.ParseFloat(r.FormValue("target_amount"), 64)
	networkTypeID, _ := strconv.Atoi(r.FormValue("network_type_id"))
	receiverWallet := r.FormValue("receiver_wallet")
	payoutConditionID, _ := strconv.Atoi(r.FormValue("payout_condition_id"))

	// 3. Xử lý lưu File
	imagePath, err := saveUploadedFile(r, "image")
	if err != nil {
		http.Error(w, "Failed to save image: "+err.Error(), http.StatusInternalServerError)
		return
	}

	idFilesPath, err := saveUploadedFile(r, "id_files")
	if err != nil {
		http.Error(w, "Failed to save ID files: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Tạo JSON tĩnh cho các trường map (nếu cần)
	idFilesJSON := fmt.Sprintf(`{"status": "uploaded", "url": "%s"}`, idFilesPath)
	linksJSON := `{}`
	tempContractAddress := fmt.Sprintf("0x00000000000000000000000000000000000%d", time.Now().Unix())

	// 4. Lưu vào Database
	query := `
		INSERT INTO projects (
			title, category_id, description, creator_wallet, image,
			beneficiary_name, beneficiary_contact, id_files,
			address, district, province,
			target_amount, network_type_id, receiver_wallet, payout_condition_id, 
			contract_address, status, links
		) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18) 
		RETURNING id`

	var newProjectID string
	err = utils.DB.QueryRow(query,
		title, categoryID, description, creatorWallet, imagePath,
		beneficiaryName, beneficiaryContact, idFilesJSON,
		address, district, province,
		targetAmount, networkTypeID, receiverWallet, payoutConditionID,
		tempContractAddress, 0, linksJSON,
	).Scan(&newProjectID)

	if err != nil {
		http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message":    "Project created successfully",
		"project_id": newProjectID,
	})
}

// -------------------------------------------------------------

type UpdateProjectRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Image       string `json:"image"`
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
			image = $3,
			status = $4, 
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $5
	`
	res, err := utils.DB.Exec(query, body.Title, body.Description, body.Image, body.Status, id)

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

// ===== HÀM XÓA DỰ ÁN (TỪ CHỐI) =====
func DeleteProjectHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Chỉ chấp nhận method DELETE
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Lấy id từ query ?id=UUID
	id := r.URL.Query().Get("id")
	if strings.TrimSpace(id) == "" {
		http.Error(w, "Missing project id", http.StatusBadRequest)
		return
	}

	// Lệnh SQL xóa vĩnh viễn khỏi Database
	query := `DELETE FROM projects WHERE id = $1`
	res, err := utils.DB.Exec(query, id)

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
		"message": "Project deleted successfully",
	})
}
