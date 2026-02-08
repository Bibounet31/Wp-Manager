package handlers

import (
	"log"
	"net/http"
)

func AddfavoriteHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/wallpapers", http.StatusSeeOther)
	log.Println("you better finish this feature bib.... (favorite..)")
}
