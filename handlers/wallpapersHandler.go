package handlers

import (
	"log"
	"net/http"
)

func WallpapersHandler(w http.ResponseWriter, r *http.Request) {
	user := getCurrentUser(r)
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Use helper to get userID with expiry check
	userID, err := getUserIDFromSession(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Get user's wallpapers
	rows, err := db.Query(`
		SELECT id, filename, original_name, uploaded_at, ispublic, toreview
		FROM wallpapers 
		WHERE user_id = ? 
		ORDER BY uploaded_at DESC`, userID)
	if err != nil {
		log.Println("Failed to query wallpapers:", err)
		http.Error(w, "Failed to load wallpapers", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var wallpapers []Wallpaper
	for rows.Next() {
		var w Wallpaper
		if err := rows.Scan(&w.ID, &w.Filename, &w.OriginalName, &w.UploadedAt, &w.IsPublic, &w.ToReview); err != nil {
			log.Println("Row scan error:", err)
			continue
		}
		wallpapers = append(wallpapers, w)
	}

	data := WallpapersPageData{
		Wallpapers: wallpapers,
		Username:   user.Username,
		IsAdmin:    user.IsAdmin,
	}
	if err := templates.ExecuteTemplate(w, "wallpapers.html", data); err != nil {
		log.Println("Template error:", err)
		http.Error(w, "Template error", http.StatusInternalServerError)
	}
}
