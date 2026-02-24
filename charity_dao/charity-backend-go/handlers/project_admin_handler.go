package handlers

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/thanhngv03/decentralized-charity-fund/charity-backend-go/services"
	"github.com/thanhngv03/decentralized-charity-fund/charity-backend-go/utils"
)

func CreateProjectHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", 405)
		return
	}

	var body struct {
		Title           string `json:"title"`
		Description     string `json:"description"`
		TargetAmountWei string `json:"target_amount_wei"`
		Category        string `json:"category"`
		StartDate       string `json:"start_date"`
		EndDate         string `json:"end_date"`
		WalletAddress   string `json:"wallet_address"`
		OwnerName       string `json:"owner_name"`
		OwnerEmail      string `json:"owner_email"`
		OwnerPhone      string `json:"owner_phone"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid body", 400)
		return
	}

	// DEPLOY SMART CONTRACT
	contractAddress, err := services.DeployProjectContract(body.TargetAmountWei)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// ===== 2️⃣ INSERT INTO DATABASE =====
	_, err = utils.DB.Exec(`
		INSERT INTO projects
		(title, description, target_amount_wei, category,
		 wallet_address, start_date, end_date,
		 owner_name, owner_email, owner_phone,
		 collected_amount_wei, status, approved)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,0,'calling',false)
	`,
		body.Title,
		body.Description,
		body.TargetAmountWei,
		body.Category,
		contractAddress.Hex(),
		body.StartDate,
		body.EndDate,
		body.OwnerName,
		body.OwnerEmail,
		body.OwnerPhone,
	)

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{
		"message":          "Project created successfully",
		"contract_address": contractAddress.Hex(),
	})
}

// CheckAdmin là một phần mềm trung gian đơn giản để bảo vệ các điểm cuối của quản trị viên.
// về sau mở rộng sau này để thực hiện xác thực/ủy quyền thực sự.
func CheckAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// #region agent log
		f, err := os.OpenFile("debug-f07305.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err == nil {
			_, _ = f.WriteString(`{"id":"log_check_admin","timestamp":0,"location":"project_admin_handler.go:CheckAdmin","message":"CheckAdmin called","data":{"path":"` + r.URL.Path + `"},"runId":"compile_fix","hypothesisId":"H1"}` + "\n")
			_ = f.Close()
		}
		// #endregion agent log

		// TODO: add real admin checks here (e.g., headers, tokens)
		next.ServeHTTP(w, r)
	}
}

func UpdateProjectHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Lấy id từ query ?id=1
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Missing project id", http.StatusBadRequest)
		return
	}

	var body struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Status      string `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	_, err := utils.DB.Exec(`
		UPDATE projects
		SET title=$1,
		    description=$2,
		    status=$3,
		    updated_at=NOW()
		WHERE id=$4
	`,
		body.Title,
		body.Description,
		body.Status,
		id,
	)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"message": "Project updated successfully",
	})
}
