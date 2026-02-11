package handlers

import (
	"log"
	"net/http"
)

func DeletewpHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/wallpapers", http.StatusSeeOther)

	userID, err := getUserIDFromSession(r)
	if err != nil {
		log.Println("Session error:", err)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	wallpaperID := r.FormValue("wallpaper_id")
	if wallpaperID == "" {
		http.Error(w, "Wallpaper ID missing", http.StatusBadRequest)
		return
	}

	//delete seleccted wp
	_, err = db.Exec("DELETE FROM wallpapers WHERE id = ?", wallpaperID)
	if err != nil {
		log.Println("failed to delete..", err)
	}

	log.Println("user", userID, "deleted a wp!")

}
