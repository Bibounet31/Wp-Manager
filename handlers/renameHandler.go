package handlers

import (
	"log"
	"net/http"
)

func RenameHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get form values
	wallpaperID := r.FormValue("wallpaper_id")
	newName := r.FormValue("new_name")

	if wallpaperID == "" || newName == "" {
		http.Error(w, "Missing wallpaper ID or new name", http.StatusBadRequest)
		return
	}
	if len(newName) > 255 {
		http.Error(w, "Name too long (max 255 characters)", http.StatusBadRequest)
		return
	}

	userID, err := getUserIDFromSession(r)
	if err != nil {
		log.Println("Session error:", err)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Verify that this wallpaper belongs to the user, to not share ur private wallpapers with someone else~
	var ownerID int
	err = db.QueryRow("SELECT user_id FROM wallpapers WHERE id = ?", wallpaperID).Scan(&ownerID)
	if err != nil {
		http.Error(w, "Wallpaper not found", http.StatusNotFound)
		return
	}

	if ownerID != userID {
		log.Printf("⚠️ Unauthorized rename attempt: user %d tried to rename wallpaper owned by %d", userID, ownerID)
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	// update wallpaper name
	_, err = db.Exec("UPDATE wallpapers SET original_name = ? WHERE id = ?", newName, wallpaperID)
	if err != nil {
		log.Println("Failed to rename wallpaper:", err)
		http.Error(w, "Failed to rename wallpaper", http.StatusInternalServerError)
		return
	}

	log.Printf("✅ Wallpaper %s renamed to: %s by user %d", wallpaperID, newName, userID)

	http.Redirect(w, r, "/wallpapers", http.StatusSeeOther)
}
