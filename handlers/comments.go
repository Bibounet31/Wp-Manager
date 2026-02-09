package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Comment struct {
	ID          int       `json:"id"`
	WallpaperID int       `json:"wallpaper_id"`
	UserID      int       `json:"user_id"`
	Username    string    `json:"username"`
	Text        string    `json:"text"`
	CreatedAt   time.Time `json:"created_at"`
}

type CommentRequest struct {
	WallpaperID int    `json:"wallpaper_id"`
	Text        string `json:"text"`
}

// GetCommentsHandler retrieves all comments for a wallpaper
func GetCommentsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract wallpaper ID from URL path
	// URL format: /api/comments/123
	path := strings.TrimPrefix(r.URL.Path, "/api/comments/")
	wallpaperID, err := strconv.Atoi(path)
	if err != nil {
		http.Error(w, "Invalid wallpaper ID", http.StatusBadRequest)
		return
	}

	// Query comments from database
	rows, err := db.Query(`
		SELECT c.id, c.wallpaper_id, c.user_id, u.username, c.text, c.created_at
		FROM comments c
		JOIN users u ON c.user_id = u.id
		WHERE c.wallpaper_id = ?
		ORDER BY c.created_at DESC
	`, wallpaperID)
	if err != nil {
		log.Println("Failed to query comments:", err)
		http.Error(w, "Failed to load comments", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var comments []Comment
	for rows.Next() {
		var c Comment
		if err := rows.Scan(&c.ID, &c.WallpaperID, &c.UserID, &c.Username, &c.Text, &c.CreatedAt); err != nil {
			log.Println("Row scan error:", err)
			continue
		}
		comments = append(comments, c)
	}

	// Return empty array instead of null if no comments
	if comments == nil {
		comments = []Comment{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(comments)
}

// PostCommentHandler creates a new comment
func PostCommentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if user is logged in
	userID, err := getUserIDFromSession(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var req CommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate input
	if req.WallpaperID <= 0 {
		http.Error(w, "Invalid wallpaper ID", http.StatusBadRequest)
		return
	}
	if len(req.Text) == 0 || len(req.Text) > 500 {
		http.Error(w, "Comment must be 1-500 characters", http.StatusBadRequest)
		return
	}

	// Verify wallpaper exists
	var wallpaperExists bool
	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM wallpapers WHERE id = ?)", req.WallpaperID).Scan(&wallpaperExists)
	if err != nil || !wallpaperExists {
		http.Error(w, "Wallpaper not found", http.StatusNotFound)
		return
	}

	// Insert comment
	result, err := db.Exec(`
		INSERT INTO comments (wallpaper_id, user_id, text, created_at)
		VALUES (?, ?, ?, NOW())
	`, req.WallpaperID, userID, req.Text)
	if err != nil {
		log.Println("Failed to insert comment:", err)
		http.Error(w, "Failed to post comment", http.StatusInternalServerError)
		return
	}

	commentID, _ := result.LastInsertId()
	log.Printf("✅ Comment posted: ID=%d, User=%d, Wallpaper=%d", commentID, userID, req.WallpaperID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"id":      commentID,
	})
}
