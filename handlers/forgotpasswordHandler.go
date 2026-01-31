package handlers

import "net/http"

func ForgotpasswordHandler(w http.ResponseWriter, r *http.Request) {
	err := templates.ExecuteTemplate(w, "forgot-password.html", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
