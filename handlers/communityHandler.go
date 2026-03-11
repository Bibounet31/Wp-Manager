package handlers

import (
	"log"
	"net/http"
)

type CategoryStats struct {
	Name  string
	Icon  string
	Desc  string
	Count int
}

type CommunityPageData struct {
	Categories []CategoryStats
	Wallpapers []Wallpaper
	Username   string
	IsAdmin    bool
}

func CommunityHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Define your categories
	categories := []CategoryStats{
		{Name: "Anime", Icon: "🌙", Desc: "Checkout all our anime wallpapers!"},
		{Name: "Landscapes", Icon: "🌸", Desc: "some grass?"},
		{Name: "other", Icon: "📁", Desc: "Miscellaneous wallpapers"},
	}

	// Get counts for each category
	for i := range categories {
		var count int
		err := db.QueryRow(`
			SELECT COUNT(*) 
			FROM wallpapers 
			WHERE category = ? AND ispublic = 1
		`, categories[i].Name).Scan(&count)

		if err != nil {
			log.Println("Failed to count wallpapers for", categories[i].Name, ":", err)
			count = 0
		}
		categories[i].Count = count
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

	data := CommunityPageData{
		Categories: categories,
		Wallpapers: wallpapers,
		Username:   username,
		IsAdmin:    isAdmin,
	}

	if err := templates.ExecuteTemplate(w, "community.html", data); err != nil {
		log.Println("Template error:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}
