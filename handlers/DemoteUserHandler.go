package handlers

import (
	"log"
	"net/http"
)

func DemoteUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form", http.StatusBadRequest)
		return
	}

	username := r.FormValue("username")
	if username == "" {
		http.Error(w, "Username missing", http.StatusBadRequest)
		return
	}

	result, err := db.Exec(`
		UPDATE users
		SET isadmin = 0
		WHERE username = ?
	`, username)

	if err != nil {
		log.Println("Database error:", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	rows, err := result.RowsAffected()
	if err != nil || rows == 0 {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	log.Printf("✅ User %s demoted by admin", username)
	http.Redirect(w, r, "/adminpannel", http.StatusSeeOther)
}
