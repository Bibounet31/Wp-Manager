package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserProfile struct {
	Username string
	Email    string
	Name     string
	Surname  string
}

type Wallpaper struct {
	ID           int
	UserID       int
	Filename     string
	OriginalName string
	FilePath     string
	UploadedAt   time.Time
}

type WallpapersPageData struct {
	Wallpapers []Wallpaper
}

// Register all HTTP routes
func registerRoutes() {
	http.HandleFunc("/", page("index.html"))
	http.HandleFunc("/community", page("community.html"))
	http.HandleFunc("/wallpapers", wallpapersHandler)
	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/profile", profileHandler)
	http.HandleFunc("/logout", logoutHandler)
	http.HandleFunc("/upload", uploadHandler)
	http.HandleFunc("/rename", renameHandler)

	// Serve uploaded files
	http.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("web/uploads"))))
}

// page renders GET-only pages
func page(file string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if err := templates.ExecuteTemplate(w, file, nil); err != nil {
			log.Println("Template error:", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
	}
}

// Handles user registration
func registerHandler(w http.ResponseWriter, r *http.Request) {
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

	// GET → show form
	templates.ExecuteTemplate(w, "register.html", nil)
}

// Handles user login
func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		r.ParseForm()
		username := r.FormValue("username")
		password := r.FormValue("password")

		var userID int
		var hashedPassword string

		err := db.QueryRow("SELECT id, password_hash FROM users WHERE username = ?", username).Scan(&userID, &hashedPassword)
		if err != nil {
			http.Error(w, "Invalid username or password", http.StatusUnauthorized)
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)); err != nil {
			http.Error(w, "Invalid username or password", http.StatusUnauthorized)
			return
		}

		// Create session
		sessionID := uuid.New().String()
		expiresAt := time.Now().Add(24 * time.Hour)

		_, _ = db.Exec("INSERT INTO sessions (id, user_id, expires_at) VALUES (?, ?, ?)", sessionID, userID, expiresAt)

		http.SetCookie(w, &http.Cookie{
			Name:     "session_id",
			Value:    sessionID,
			Expires:  expiresAt,
			Path:     "/",
			HttpOnly: true,
		})

		http.Redirect(w, r, "/profile", http.StatusSeeOther)
		return
	}

	// GET → show login form
	templates.ExecuteTemplate(w, "login.html", nil)
}

func renameHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get form values
	wallpaperID := r.FormValue("wallpaper_id")
	newName := r.FormValue("new_name")

	// Validate input
	if wallpaperID == "" || newName == "" {
		http.Error(w, "Missing wallpaper ID or new name", http.StatusBadRequest)
		return
	}

	// Check if user is logged in
	cookie, err := r.Cookie("session_id")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Get user ID from session
	var userID int
	err = db.QueryRow("SELECT user_id FROM sessions WHERE id = ?", cookie.Value).Scan(&userID)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Verify that this wallpaper belongs to the user (IMPORTANT for security!)
	var ownerID int
	err = db.QueryRow("SELECT user_id FROM wallpapers WHERE id = ?", wallpaperID).Scan(&ownerID)
	if err != nil {
		http.Error(w, "Wallpaper not found", http.StatusNotFound)
		return
	}

	if ownerID != userID {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	// Update the wallpaper name
	_, err = db.Exec("UPDATE wallpapers SET original_name = ? WHERE id = ?", newName, wallpaperID)
	if err != nil {
		log.Println("Failed to rename wallpaper:", err)
		http.Error(w, "Failed to rename wallpaper", http.StatusInternalServerError)
		return
	}

	log.Printf("✅ Wallpaper %s renamed to: %s by user %d", wallpaperID, newName, userID)

	// Redirect back to wallpapers page
	http.Redirect(w, r, "/wallpapers", http.StatusSeeOther)
}

// Print all users in console
func printAllUsers() {
	rows, err := db.Query("SELECT id, username, email, name, surname, created_at FROM users")
	if err != nil {
		log.Println("Failed to query users:", err)
		return
	}
	defer rows.Close()

	fmt.Println("📋 Current users in the database:")
	for rows.Next() {
		var id int
		var username, email, name, surname string
		var createdAt time.Time
		if err := rows.Scan(&id, &username, &email, &name, &surname, &createdAt); err != nil {
			log.Println("Row scan error:", err)
			continue
		}
		fmt.Printf("ID: %d | Username: %s | Email: %s | Name: %s %s | Created: %s\n",
			id, username, email, name, surname, createdAt.Format("2006-01-02 15:04:05"))
	}
}

// profileHandler shows user info if logged in
func profileHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("📍 Profile handler called")

	// Check if session cookie exists
	cookie, err := r.Cookie("session_id")
	if err != nil {
		log.Println("❌ No session cookie found")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	sessionID := cookie.Value
	log.Println("🔑 Session ID:", sessionID)

	// Get user ID from session
	var userID int
	var expiresAt time.Time
	err = db.QueryRow("SELECT user_id, expires_at FROM sessions WHERE id = ?", sessionID).
		Scan(&userID, &expiresAt)
	if err != nil {
		log.Println("❌ Session not found in DB:", err)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Check if session expired
	if time.Now().After(expiresAt) {
		log.Println("⏰ Session expired")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Get user info
	var user UserProfile
	err = db.QueryRow("SELECT username, email, name, surname FROM users WHERE id = ?", userID).
		Scan(&user.Username, &user.Email, &user.Name, &user.Surname)
	if err != nil {
		log.Println("❌ Failed to get user info:", err)
		http.Error(w, "Failed to get user info", http.StatusInternalServerError)
		return
	}

	log.Printf("✅ Loaded user: %s (%s)", user.Username, user.Email)

	// Render profile page with struct
	if err := templates.ExecuteTemplate(w, "profile.html", user); err != nil {
		log.Println("❌ Profile template error:", err)
		http.Error(w, "Template error", http.StatusInternalServerError)
	}
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_id")
	if err == nil {
		db.Exec("DELETE FROM sessions WHERE id = ?", cookie.Value)
		cookie.MaxAge = -1
		http.SetCookie(w, cookie)
	}
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// getCurrentUser returns the logged-in user or nil if not logged in
func getCurrentUser(r *http.Request) *UserProfile {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		return nil
	}

	var userID int
	var expiresAt time.Time
	err = db.QueryRow("SELECT user_id, expires_at FROM sessions WHERE id = ?", cookie.Value).
		Scan(&userID, &expiresAt)
	if err != nil || time.Now().After(expiresAt) {
		return nil
	}

	var user UserProfile
	err = db.QueryRow("SELECT username, email, name, surname FROM users WHERE id = ?", userID).
		Scan(&user.Username, &user.Email, &user.Name, &user.Surname)
	if err != nil {
		return nil
	}

	return &user
}

// wallpapersHandler displays user's wallpapers
func wallpapersHandler(w http.ResponseWriter, r *http.Request) {
	// Check if user is logged in
	user := getCurrentUser(r)
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Get user ID from session
	cookie, _ := r.Cookie("session_id")
	var userID int
	db.QueryRow("SELECT user_id FROM sessions WHERE id = ?", cookie.Value).Scan(&userID)

	// Get user's wallpapers
	rows, err := db.Query(`
		SELECT id, filename, original_name, uploaded_at 
		FROM wallpapers 
		WHERE user_id = ? 
		ORDER BY uploaded_at DESC
	`, userID)
	if err != nil {
		log.Println("Failed to query wallpapers:", err)
		http.Error(w, "Failed to load wallpapers", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var wallpapers []Wallpaper
	for rows.Next() {
		var w Wallpaper
		if err := rows.Scan(&w.ID, &w.Filename, &w.OriginalName, &w.UploadedAt); err != nil {
			log.Println("Row scan error:", err)
			continue
		}
		wallpapers = append(wallpapers, w)
	}

	data := WallpapersPageData{Wallpapers: wallpapers}
	if err := templates.ExecuteTemplate(w, "wallpapers.html", data); err != nil {
		log.Println("Template error:", err)
		http.Error(w, "Template error", http.StatusInternalServerError)
	}
}

// uploadHandler handles wallpaper uploads
func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if user is logged in
	cookie, err := r.Cookie("session_id")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Get user ID
	var userID int
	err = db.QueryRow("SELECT user_id FROM sessions WHERE id = ?", cookie.Value).Scan(&userID)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Parse multipart form (32MB max)
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, "File too large", http.StatusBadRequest)
		return
	}

	// Get file from form
	file, header, err := r.FormFile("wallpaper")
	if err != nil {
		http.Error(w, "Failed to read file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Create uploads directory if it doesn't exist
	os.MkdirAll("web/uploads", 0755)

	// Generate unique filename
	ext := filepath.Ext(header.Filename)
	filename := fmt.Sprintf("%d_%s%s", userID, uuid.New().String(), ext)
	filePath := filepath.Join("web/uploads", filename)

	// Save file
	dst, err := os.Create(filePath)
	if err != nil {
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}

	// Save to database
	_, err = db.Exec(`
		INSERT INTO wallpapers (user_id, filename, original_name, file_path)
		VALUES (?, ?, ?, ?)
	`, userID, filename, header.Filename, filePath)
	if err != nil {
		log.Println("Failed to save wallpaper to DB:", err)
		http.Error(w, "Failed to save wallpaper", http.StatusInternalServerError)
		return
	}

	log.Printf("✅ Wallpaper uploaded: %s by user %d", header.Filename, userID)
	http.Redirect(w, r, "/wallpapers", http.StatusSeeOther)
}
