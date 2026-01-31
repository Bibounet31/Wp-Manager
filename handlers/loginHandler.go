package handlers

import (
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		r.ParseForm()
		username := r.FormValue("username")
		password := r.FormValue("password")

		var userID int
		var hashedPassword string
		var isAdmin bool

		// Fetch user data from db
		err := db.QueryRow("SELECT id, password_hash, isadmin FROM users WHERE username = ?", username).
			Scan(&userID, &hashedPassword, &isAdmin)
		if err != nil {
			http.Error(w, "Invalid username or password", http.StatusUnauthorized)
			return
		}

		// Verify password
		if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)); err != nil {
			http.Error(w, "Invalid username or password", http.StatusUnauthorized)
			return
		}

		// Create session wooo
		sessionID := uuid.New().String()
		expiresAt := time.Now().Add(24 * time.Hour)

		_, _ = db.Exec("INSERT INTO sessions (id, user_id, expires_at) VALUES (?, ?, ?)", sessionID, userID, expiresAt)

		// Set session cookie (yum)
		http.SetCookie(w, &http.Cookie{
			Name:     "session_id",
			Value:    sessionID,
			Expires:  expiresAt,
			Path:     "/",
			HttpOnly: true,
			SameSite: http.SameSiteStrictMode,
		})

		// Log the login and admin status
		log.Printf("username: %s successfully logged in, isadmin: %t", username, isAdmin)

		http.Redirect(w, r, "/profile", http.StatusSeeOther)
		return
	}

	templates.ExecuteTemplate(w, "login.html", nil)
}
