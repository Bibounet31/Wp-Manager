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
	log.Println("📥 GetCommentsHandler called")

	if r.Method != http.MethodGet {
		log.Println("❌ Wrong method:", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract wallpaper ID from URL path
	// URL format: /api/comments/123
	path := strings.TrimPrefix(r.URL.Path, "/api/comments/")
	log.Println("📝 Path:", path)

	wallpaperID, err := strconv.Atoi(path)
	if err != nil {
		log.Println("❌ Invalid wallpaper ID:", path, err)
		http.Error(w, "Invalid wallpaper ID", http.StatusBadRequest)
		return
	}

	log.Println("🔍 Querying comments for wallpaper:", wallpaperID)

	// Query comments from database
	rows, err := db.Query(`
       SELECT c.id, c.wallpaper_id, c.user_id, u.username, c.text, c.created_at
       FROM comments c
       JOIN users u ON c.user_id = u.id
       WHERE c.wallpaper_id = ?
       ORDER BY c.created_at DESC
    `, wallpaperID)
	if err != nil {
		log.Println("❌ Failed to query comments:", err)
		http.Error(w, "Failed to load comments", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var comments []Comment
	for rows.Next() {
		var c Comment
		if err := rows.Scan(&c.ID, &c.WallpaperID, &c.UserID, &c.Username, &c.Text, &c.CreatedAt); err != nil {
			log.Println("❌ Row scan error:", err)
			continue
		}
		comments = append(comments, c)
	}

	// Return empty array if no comments
	if comments == nil {
		comments = []Comment{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(comments)
}

// PostCommentHandler creates a new comment
func PostCommentHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("PostCommentHandler called")
	log.Println("Method:", r.Method)
	log.Println("Path:", r.URL.Path)

	// Log all cookies
	log.Println("🍪 Cookies received:")
	for _, cookie := range r.Cookies() {
		log.Printf("  - %s = %s\n", cookie.Name, cookie.Value[:min(len(cookie.Value), 20)])
	}

	if r.Method != http.MethodPost {
		log.Println("Wrong method:", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if user is logged in
	log.Println("🔐 Checking user session...")
	userID, err := getUserIDFromSession(r)
	if err != nil {
		log.Println("❌ Not authorized:", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"error":   "Please log in to post comments",
			"details": err.Error(),
		})
		return
	}

	// Parse request body
	var req CommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	log.Printf("Request: WallpaperID=%d, Text length=%d\n", req.WallpaperID, len(req.Text))

	// Validate input
	if req.WallpaperID <= 0 {
		log.Println("❌ Invalid wallpaper ID:", req.WallpaperID)
		http.Error(w, "Invalid wallpaper ID", http.StatusBadRequest)
		return
	}
	if len(req.Text) == 0 || len(req.Text) > 500 {
		log.Printf("❌ Invalid text length: %d\n", len(req.Text))
		http.Error(w, "Comment must be 1-500 characters", http.StatusBadRequest)
		return
	}

	// Verify wallpaper exists
	log.Println("🔍 Checking if wallpaper exists...")
	var wallpaperExists bool
	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM wallpapers WHERE id = ?)", req.WallpaperID).Scan(&wallpaperExists)
	if err != nil {
		log.Println("❌ Database error checking wallpaper:", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	if !wallpaperExists {
		log.Println("❌ Wallpaper not found:", req.WallpaperID)
		http.Error(w, "Wallpaper not found", http.StatusNotFound)
		return
	}
	log.Println("Wallpaper exists")

	// Insert comment
	log.Println("Inserting comment into database...")
	result, err := db.Exec(`
       INSERT INTO comments (wallpaper_id, user_id, text, created_at)
       VALUES (?, ?, ?, NOW())
    `, req.WallpaperID, userID, req.Text)
	if err != nil {
		log.Println("❌ Failed to insert comment:", err)
		http.Error(w, "Failed to post comment", http.StatusInternalServerError)
		return
	}

	commentID, _ := result.LastInsertId()
	log.Printf("Comment posted successfully: ID=%d, User=%d, Wallpaper=%d\n", commentID, userID, req.WallpaperID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"id":      commentID,
	})
}
