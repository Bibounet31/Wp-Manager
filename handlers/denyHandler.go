package handlers

import (
	"database/sql"
	"log"
	"net/http"
)

func DenyHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("DenyHandler called :3")
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse form data
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form", http.StatusBadRequest)
		return
	}

	wallpaperID := r.FormValue("wallpaper_id")
	if wallpaperID == "" {
		http.Error(w, "Wallpaper ID missing", http.StatusBadRequest)
		return
	}

	// Get userid
	userID, err := getUserIDFromSession(r)
	if err != nil {
		log.Println("Session error:", err)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// check if wallpaper is owned by user + get current toreview status
	var ownerID int
	var toReview bool
	err = db.QueryRow("SELECT user_id, COALESCE(toreview, 0) FROM wallpapers WHERE id = ?", wallpaperID).
		Scan(&ownerID, &toReview)
	if err != nil {
		log.Println("Wallpaper not found:", err)
		http.Error(w, "Wallpaper not found", http.StatusNotFound)
		return
	}

	user := getCurrentUser(r)
	isAdmin := user != nil && user.IsAdmin

	if !isAdmin && ownerID != userID {
		log.Printf("⚠️ Unauthorized toggle attempt: user %d tried to toggle wallpaper owned by %d", userID, ownerID)
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}
	// Toggle the toReview status
	var result sql.Result
	if toReview {
		// if ==1, make it 0
		result, err = db.Exec("UPDATE wallpapers SET toreview = 0 WHERE id = ?", wallpaperID)
		if err != nil {
			log.Println("Failed to unpublish wallpaper:", err)
			http.Error(w, "Failed to unpublish wallpaper", http.StatusInternalServerError)
			return
		}
		log.Printf("Wallpaper %s unpublished by user %d", wallpaperID, userID)
	} else {
		// if ==0, make it 1
		result, err = db.Exec("UPDATE wallpapers SET toreview = 0 WHERE id = ?", wallpaperID)
		if err != nil {
			log.Println("Failed to publish wallpaper:", err)
			http.Error(w, "Failed to publish wallpaper", http.StatusInternalServerError)
			return
		}
		log.Printf("Wallpaper %s published by user %d", wallpaperID, userID)
	}

	rows, err := result.RowsAffected()
	if err != nil || rows == 0 {
		http.Error(w, "Wallpaper not found", http.StatusNotFound)
		return
	}

	http.Redirect(w, r, "/adminpannel", http.StatusSeeOther)
}
