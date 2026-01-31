package handlers

import (
	"log"
	"net/http"
)

func ProfileHandler(w http.ResponseWriter, r *http.Request) {
	user := getCurrentUser(r) // checks expiry
	if user == nil {
		log.Println("❌ No valid session")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Render profile page with struct
	if err := templates.ExecuteTemplate(w, "profile.html", user); err != nil {
		log.Println("❌ Profile template error:", err)
		http.Error(w, "Template error", http.StatusInternalServerError)
	}
}
