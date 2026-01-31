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

// structs:
type UserProfile struct {
	Username string
	Email    string
	Name     string
	Surname  string
	IsAdmin  bool
	UserID   int
}

type AdminPanelData struct {
	CurrentUser *UserProfile
	AllUsers    []UserProfile
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
	Username   string
	IsAdmin    bool
}

type PageData struct {
	Username string
	IsAdmin  bool
}

// HTTP ROUTES
func registerRoutes() {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/community", communityHandler)
	http.HandleFunc("/wallpapers", wallpapersHandler)
	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/profile", profileHandler)
	http.HandleFunc("/logout", logoutHandler)
	http.HandleFunc("/upload", uploadHandler)
	http.HandleFunc("/rename", renameHandler)
	http.HandleFunc("/adminpannel", adminpannelHandler)
	http.HandleFunc("/admin/promote", PromoteUserHandler)
	http.HandleFunc("/admin/demote", DemoteUserHandler)
	http.HandleFunc("/admin/deleteacc", DeleteAccHandler)
	http.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("web/uploads"))))
}

// get page data with user infos
func getPageData(r *http.Request) PageData {
	data := PageData{}
	user := getCurrentUser(r)
	if user != nil {
		data.Username = user.Username
		data.IsAdmin = user.IsAdmin
	}
	return data
}

func DeleteAccHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// check if POST
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form", http.StatusBadRequest)
		return
	}

	userID := r.FormValue("user_id")
	if userID == "" {
		http.Error(w, "User ID missing", http.StatusBadRequest)
		return
	}

	// Prevent deleting yourself, that would suck
	currentUserID := r.Context().Value("userid") // Make sure this returns a string
	if currentUserID != nil && fmt.Sprintf("%v", currentUserID) == userID {
		http.Error(w, "You cannot delete yourself", http.StatusForbidden)
		return
	}
	//SQL stuff deleting
	result, err := db.Exec(`
        DELETE FROM users
        WHERE id = ?
    `, userID)

	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	rows, err := result.RowsAffected()
	if err != nil || rows == 0 {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Redirect back to admin panel
	http.Redirect(w, r, "/adminpannel", http.StatusSeeOther)
}

// goes from admin > user
func DemoteUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// check if POST
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form", http.StatusBadRequest)
		return
	}

	username := r.FormValue("username")

	// Execute update
	result, err := db.Exec(`
        UPDATE users
        SET isadmin = 0
        WHERE username = ?
    `, username)

	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	rows, err := result.RowsAffected()
	if err != nil || rows == 0 {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	http.Redirect(w, r, "/adminpannel", http.StatusSeeOther)
}

// promote a user (like it wasn't obvious)
func PromoteUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// check if post
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form", http.StatusBadRequest)
		return
	}

	username := r.FormValue("username")
	if username == "" {
		http.Error(w, "Username missing", http.StatusBadRequest)
		return
	}

	// Execute update
	result, err := db.Exec(`
        UPDATE users
        SET isadmin = 1
        WHERE username = ?
    `, username)

	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	rows, err := result.RowsAffected()
	if err != nil || rows == 0 {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Redirect back to admin panel
	http.Redirect(w, r, "/adminpannel", http.StatusSeeOther)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	data := getPageData(r)
	if err := templates.ExecuteTemplate(w, "index.html", data); err != nil {
		log.Println("Template error:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func adminpannelHandler(w http.ResponseWriter, r *http.Request) {
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

	// Query all users from database - ADD id field
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
		// ADD &u.UserID to the scan
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

func communityHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	data := getPageData(r)
	if err := templates.ExecuteTemplate(w, "community.html", data); err != nil {
		log.Println("Template error:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
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

// Handles user login
func loginHandler(w http.ResponseWriter, r *http.Request) {
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

		// Redirect to profile
		http.Redirect(w, r, "/profile", http.StatusSeeOther)
		return
	}

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

	if wallpaperID == "" || newName == "" {
		http.Error(w, "Missing wallpaper ID or new name", http.StatusBadRequest)
		return
	}
	if len(newName) > 255 {
		http.Error(w, "Name too long (max 255 characters)", http.StatusBadRequest)
		return
	}

	// ✅ Use getUserIDFromSession helper with expiry check
	userID, err := getUserIDFromSession(r)
	if err != nil {
		log.Println("Session error:", err)
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
		log.Printf("⚠️ Unauthorized rename attempt: user %d tried to rename wallpaper owned by %d", userID, ownerID)
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
	// ✅ Use getCurrentUser which checks expiry
	user := getCurrentUser(r)
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

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_id")
	if err == nil {
		db.Exec("DELETE FROM sessions WHERE id = ?", cookie.Value)
		cookie.MaxAge = -1
		http.SetCookie(w, cookie)
	}
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// ✅ NEW: Helper function to get user ID from session WITH expiry check
func getUserIDFromSession(r *http.Request) (int, error) {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		return 0, fmt.Errorf("no session cookie")
	}

	var userID int
	var expiresAt time.Time
	err = db.QueryRow("SELECT user_id, expires_at FROM sessions WHERE id = ?", cookie.Value).
		Scan(&userID, &expiresAt)
	if err != nil {
		return 0, fmt.Errorf("session not found: %w", err)
	}

	// ✅ Check if session expired
	if time.Now().After(expiresAt) {
		// Clean up expired session
		db.Exec("DELETE FROM sessions WHERE id = ?", cookie.Value)
		return 0, fmt.Errorf("session expired")
	}

	return userID, nil
}

// getCurrentUser returns the logged-in user or nil if not logged in
func getCurrentUser(r *http.Request) *UserProfile {
	userID, err := getUserIDFromSession(r)
	if err != nil {
		return nil
	}

	var user UserProfile
	err = db.QueryRow("SELECT username, email, name, surname, isadmin FROM users WHERE id = ?", userID).
		Scan(&user.Username, &user.Email, &user.Name, &user.Surname, &user.IsAdmin)
	if err != nil {
		return nil
	}

	return &user
}

func wallpapersHandler(w http.ResponseWriter, r *http.Request) {
	// ✅ Use getCurrentUser which now checks expiry
	user := getCurrentUser(r)
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// ✅ Use helper to get userID with expiry check
	userID, err := getUserIDFromSession(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

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

	data := WallpapersPageData{
		Wallpapers: wallpapers,
		Username:   user.Username,
		IsAdmin:    user.IsAdmin,
	}
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

	// ✅ Use helper with expiry check
	userID, err := getUserIDFromSession(r)
	if err != nil {
		log.Println("Upload: Session error:", err)
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

	// ✅ Validate file extension
	ext := filepath.Ext(header.Filename)
	allowedExts := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true}
	if !allowedExts[ext] {
		http.Error(w, "Invalid file type. Only images allowed", http.StatusBadRequest)
		return
	}

	// ✅ Validate filename length
	if len(header.Filename) > 255 {
		http.Error(w, "Filename too long", http.StatusBadRequest)
		return
	}

	// Create uploads directory if it doesn't exist
	os.MkdirAll("web/uploads", 0755)

	// Generate unique filename
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
