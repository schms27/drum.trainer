package auth

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
)

const stateCookieName = "oauth_state"

// GenerateState creates a random CSRF state value, stores it in a short-lived
// HttpOnly cookie, and returns the value to embed in the OAuth redirect URL.
func GenerateState(w http.ResponseWriter, devMode bool) (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	state := hex.EncodeToString(b)
	http.SetCookie(w, &http.Cookie{
		Name:     stateCookieName,
		Value:    state,
		Path:     "/",
		MaxAge:   600, // 10 minutes — enough for the OAuth round-trip
		HttpOnly: true,
		Secure:   !devMode,
		SameSite: http.SameSiteLaxMode,
	})
	return state, nil
}

// VerifyState checks that the state query parameter matches the state cookie.
func VerifyState(r *http.Request, state string) bool {
	if state == "" {
		return false
	}
	cookie, err := r.Cookie(stateCookieName)
	if err != nil {
		return false
	}
	return cookie.Value == state
}

// ClearState expires the state cookie after a completed (or failed) callback.
func ClearState(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:   stateCookieName,
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
}
