package models

import "time"

type NewsPost struct {
	ID            int       `json:"id"`
	ProjectID     string    `json:"project_id"`
	UpdateContent string    `json:"update_content"`
	CreatedAt     time.Time `json:"created_at"`
	// Các trường JOIN từ bảng projects để hiển thị lên UI
	ProjectTitle  string `json:"project_title"`
	ProjectImage  string `json:"project_image"`
	LikesCount    int    `json:"likes_count"`
	CommentsCount int    `json:"comments_count"`
}

type NewsComment struct {
	ID            int       `json:"id"`
	NewsPostID    int       `json:"news_post_id"`
	WalletAddress string    `json:"wallet_address"`
	Content       string    `json:"content"`
	CreatedAt     time.Time `json:"created_at"`
}
