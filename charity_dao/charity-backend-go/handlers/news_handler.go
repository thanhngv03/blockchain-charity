package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/thanhngv03/decentralized-charity-fund/charity-backend-go/models"
	"github.com/thanhngv03/decentralized-charity-fund/charity-backend-go/utils"
)

// API 1: Lấy danh sách bài viết
func GetNewsFeed(w http.ResponseWriter, r *http.Request) {
	query := `
        SELECT 
            n.id, n.project_id, n.update_content, n.created_at,
            p.title, p.image,
            (SELECT COUNT(*) FROM news_likes WHERE news_post_id = n.id) as likes_count,
            (SELECT COUNT(*) FROM news_comments WHERE news_post_id = n.id) as comments_count
        FROM news_posts n
        JOIN projects p ON n.project_id = p.id
        ORDER BY n.created_at DESC
    `
	rows, err := utils.DB.Query(query)
	if err != nil {
		http.Error(w, "Lỗi truy vấn: "+err.Error(), 500)
		return
	}
	defer rows.Close()

	var posts []models.NewsPost
	for rows.Next() {
		var p models.NewsPost
		err := rows.Scan(&p.ID, &p.ProjectID, &p.UpdateContent, &p.CreatedAt, &p.ProjectTitle, &p.ProjectImage, &p.LikesCount, &p.CommentsCount)
		if err == nil {
			posts = append(posts, p)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(posts)
}

// API 2: Xử lý Like / Unlike
func ToggleLikeNews(w http.ResponseWriter, r *http.Request) {
	var req struct {
		NewsPostID    int    `json:"news_post_id"`
		WalletAddress string `json:"wallet_address"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Dữ liệu không hợp lệ", http.StatusBadRequest)
		return
	}

	var exists bool
	checkQuery := `SELECT EXISTS(SELECT 1 FROM news_likes WHERE news_post_id=$1 AND wallet_address=$2)`
	utils.DB.QueryRow(checkQuery, req.NewsPostID, req.WalletAddress).Scan(&exists)

	w.Header().Set("Content-Type", "application/json")
	if exists {
		utils.DB.Exec(`DELETE FROM news_likes WHERE news_post_id=$1 AND wallet_address=$2`, req.NewsPostID, req.WalletAddress)
		json.NewEncoder(w).Encode(map[string]string{"message": "Unliked"})
	} else {
		utils.DB.Exec(`INSERT INTO news_likes (news_post_id, wallet_address) VALUES ($1, $2)`, req.NewsPostID, req.WalletAddress)
		json.NewEncoder(w).Encode(map[string]string{"message": "Liked"})
	}
}

// API 3: Đăng Bình luận
func AddNewsComment(w http.ResponseWriter, r *http.Request) {
	var req models.NewsComment
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Dữ liệu không hợp lệ", http.StatusBadRequest)
		return
	}

	query := `INSERT INTO news_comments (news_post_id, wallet_address, content) VALUES ($1, $2, $3) RETURNING id, created_at`
	err := utils.DB.QueryRow(query, req.NewsPostID, req.WalletAddress, req.Content).Scan(&req.ID, &req.CreatedAt)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(req)
}

// API 4: Đăng bài mới (Admin)
func CreateNewsPost(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ProjectID     string `json:"project_id"` // Nhận UUID dưới dạng string
		UpdateContent string `json:"update_content"`
	}

	// Giải mã JSON một lần duy nhất
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON không hợp lệ", http.StatusBadRequest)
		return
	}

	// Kiểm tra dữ liệu đầu vào
	if req.ProjectID == "" || req.UpdateContent == "" {
		http.Error(w, "Thiếu ProjectID hoặc nội dung cập nhật", http.StatusBadRequest)
		return
	}

	// Câu lệnh SQL: PostgreSQL sẽ tự hiểu chuỗi này là UUID nếu cột tương ứng là kiểu UUID
	query := `INSERT INTO news_posts (project_id, update_content) VALUES ($1, $2) RETURNING id`
	var id int
	// Truyền trực tiếp req.ProjectID (chuỗi UUID) vào Query
	err := utils.DB.QueryRow(query, req.ProjectID, req.UpdateContent).Scan(&id)

	if err != nil {
		http.Error(w, "Lỗi DB: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"post_id": id})
}

// API 5: Lấy danh sách bình luận của 1 bài post
func GetNewsComments(w http.ResponseWriter, r *http.Request) {
	postID := r.URL.Query().Get("post_id")
	if postID == "" {
		http.Error(w, "Thiếu post_id", http.StatusBadRequest)
		return
	}

	query := `
        SELECT id, news_post_id, wallet_address, content, created_at 
        FROM news_comments 
        WHERE news_post_id = $1 
        ORDER BY created_at ASC
    `
	rows, err := utils.DB.Query(query, postID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var comments []models.NewsComment
	for rows.Next() {
		var c models.NewsComment
		err := rows.Scan(&c.ID, &c.NewsPostID, &c.WalletAddress, &c.Content, &c.CreatedAt)
		if err == nil {
			comments = append(comments, c)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(comments)
}
