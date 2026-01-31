package handlers

import (
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}

		username := r.FormValue("username")
		email := r.FormValue("mail")
		name := r.FormValue("name")
		surname := r.FormValue("surname")
		password := r.FormValue("password")

		// Input validation
		if len(username) < 3 || len(username) > 50 {
			http.Error(w, "Username must be 3-50 characters", http.StatusBadRequest)
			return
		}
		if len(password) < 8 {
			http.Error(w, "Password must be at least 8 characters", http.StatusBadRequest)
			return
		}
		if len(email) > 100 {
			http.Error(w, "Email too long", http.StatusBadRequest)
			return
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "Failed to hash password", http.StatusInternalServerError)
			return
		}

		_, err = db.Exec(`
			INSERT INTO users (username, email, name, surname, password_hash)
			VALUES (?, ?, ?, ?, ?)`,
			username, email, name, surname, hashedPassword,
		)
		if err != nil {
			http.Error(w, "Failed to save user: "+err.Error(), http.StatusInternalServerError)
			return
		}

		printAllUsers()

		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	templates.ExecuteTemplate(w, "register.html", nil)
}
