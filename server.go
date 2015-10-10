package main

import (
	"log"
	"net/http"

	"golang.org/x/oauth2"
)

const (
	HTMLIndex    = `<html><body>Logged in with <a href="/login">Facebook</a></body></html>`
	CallbackPage = `<html><body>Access Token has been updated</body></html>`
)

// RunHTTPServer Run HTTP Server for serving new access token key
func RunHTTPServer() {
	http.HandleFunc("/", handleMain)
	http.HandleFunc("/login", handleFacebookLogin)
	http.HandleFunc("/facebook_cb", handleFacebookCallback)
	http.ListenAndServe(":3000", nil)
}

func handleMain(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(HTMLIndex))
}

func handleFacebookLogin(w http.ResponseWriter, r *http.Request) {
	url := oauthConf.AuthCodeURL(oauthStateString, oauth2.AccessTypeOnline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func handleFacebookCallback(w http.ResponseWriter, r *http.Request) {
	state := r.FormValue("state")
	if state != oauthStateString {
		log.Printf("invalid oauth state, expected '%s', got '%s'", oauthStateString, state)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	}

	code := r.FormValue("code")
	token, oauthErr := oauthConf.Exchange(oauth2.NoContext, code)
	if oauthErr != nil {
		log.Printf("oauthConf.Exchange() failed with '%s'", oauthErr)
	}
	saveErr := SaveToken(token)
	if saveErr != nil {
		log.Printf("Unable to save token '%s'", saveErr)
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(CallbackPage))
}
