package handlers

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

func UploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// expiry check
	userID, err := getUserIDFromSession(r)
	if err != nil {
		log.Println("Upload: Session error:", err)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Get file from form
	file, header, err := r.FormFile("wallpaper")
	if err != nil {
		http.Error(w, "Failed to read file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// check file extension
	ext := filepath.Ext(header.Filename)
	allowedExts := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true}
	if !allowedExts[ext] {
		http.Error(w, "Invalid file type. Only images allowed", http.StatusBadRequest)
		return
	}

	// check filename length
	if len(header.Filename) > 255 {
		http.Error(w, "Filename too long", http.StatusBadRequest)
		return
	}

	// Create uploads folder if it doesn't exist
	os.MkdirAll("web/uploads", 0755)

	// Generate random filename
	filename := fmt.Sprintf("%d_%s%s", userID, uuid.New().String(), ext)
	filePath := filepath.Join("web/uploads", filename)

	// Save file
	dst, err := os.Create(filePath)
	if err != nil {
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}

	// Save to database
	_, err = db.Exec(`
		INSERT INTO wallpapers (user_id, filename, original_name, file_path)
		VALUES (?, ?, ?, ?)
	`, userID, filename, header.Filename, filePath)
	if err != nil {
		log.Println("Failed to save wallpaper to DB:", err)
		http.Error(w, "Failed to save wallpaper", http.StatusInternalServerError)
		return
	}

	log.Printf("✅ Wallpaper uploaded: %s by user %d", header.Filename, userID)
	http.Redirect(w, r, "/wallpapers", http.StatusSeeOther)
}
