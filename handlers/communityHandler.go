package handlers

import (
	"log"
	"net/http"
)

func CommunityHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get all public wallpapers
	rows, err := db.Query(`
       SELECT id, filename, original_name, uploaded_at, ispublic
       FROM wallpapers 
       WHERE ispublic = 1
       ORDER BY uploaded_at DESC
    `)
	if err != nil {
		log.Println("Failed to query wallpapers:", err)
		http.Error(w, "Failed to load wallpapers", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Add all wallpapers to the list
	var wallpapers []Wallpaper
	for rows.Next() {
		var w Wallpaper
		if err := rows.Scan(&w.ID, &w.Filename, &w.OriginalName, &w.UploadedAt, &w.IsPublic); err != nil {
			log.Println("Row scan error:", err)
			continue
		}
		wallpapers = append(wallpapers, w)
	}

	// Get current user info (if logged in)
	user := getCurrentUser(r)
	username := ""
	isAdmin := false
	if user != nil {
		username = user.Username
		isAdmin = user.IsAdmin
	}

	data := WallpapersPageData{
		Wallpapers: wallpapers,
		Username:   username,
		IsAdmin:    isAdmin,
	}

	if err := templates.ExecuteTemplate(w, "community.html", data); err != nil {
		log.Println("Template error:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}
