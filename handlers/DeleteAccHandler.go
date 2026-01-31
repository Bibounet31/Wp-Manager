package handlers

import (
	"fmt"
	"log"
	"net/http"
)

func DeleteAccHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form", http.StatusBadRequest)
		return
	}

	userID := r.FormValue("user_id")
	if userID == "" {
		http.Error(w, "User ID missing", http.StatusBadRequest)
		return
	}

	currentUserID, err := getUserIDFromSession(r)
	if err != nil {
		log.Println("Session error:", err)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if fmt.Sprintf("%d", currentUserID) == userID {
		http.Error(w, "You cannot delete yourself", http.StatusForbidden)
		return
	}

	result, err := db.Exec(`
		DELETE FROM users
		WHERE id = ?
	`, userID)

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

	log.Printf("✅ User %s deleted by admin", userID)
	http.Redirect(w, r, "/adminpannel", http.StatusSeeOther)
}
