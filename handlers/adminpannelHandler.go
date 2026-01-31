package handlers

import (
	"log"
	"net/http"
)

func AdminpannelHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get current user
	user := getCurrentUser(r)
	if user == nil {
		log.Println("❌ No valid session")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Check if user is admin
	if !user.IsAdmin {
		log.Printf("⚠️ Non-admin user %s tried to access admin panel", user.Username)
		http.Error(w, "Forbidden: Admin access required", http.StatusForbidden)
		return
	}

	// get all users from db
	rows, err := db.Query("SELECT id, username, email, name, surname, isadmin FROM users ORDER BY username")
	if err != nil {
		log.Println("Failed to query users:", err)
		http.Error(w, "Failed to load users", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var allUsers []UserProfile
	for rows.Next() {
		var u UserProfile
		if err := rows.Scan(&u.UserID, &u.Username, &u.Email, &u.Name, &u.Surname, &u.IsAdmin); err != nil {
			log.Println("Row scan error:", err)
			continue
		}
		allUsers = append(allUsers, u)
	}

	// Prepare data for template
	data := AdminPanelData{
		CurrentUser: user,
		AllUsers:    allUsers,
	}

	if err := templates.ExecuteTemplate(w, "adminpannel.html", data); err != nil {
		log.Println("Template error:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}
