package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"
)

const stateCookieName = "oauth_state"

// GenerateState creates a random CSRF state value, stores it in a short-lived
// HttpOnly cookie signed with secret, and returns the value to embed in the
// OAuth redirect URL. The signature stops an attacker who can plant cookies
// (e.g. from a sibling subdomain) from forging a matching state pair.
func GenerateState(w http.ResponseWriter, secret string, devMode bool) (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	state := hex.EncodeToString(b)
	http.SetCookie(w, &http.Cookie{
		Name:     stateCookieName,
		Value:    state + "." + signState(state, secret),
		Path:     "/",
		MaxAge:   600, // 10 minutes — enough for the OAuth round-trip
		HttpOnly: true,
		Secure:   !devMode,
		SameSite: http.SameSiteLaxMode,
	})
	return state, nil
}

// VerifyState checks that the state cookie carries a valid signature and that
// its value matches the state query parameter.
func VerifyState(r *http.Request, state, secret string) bool {
	if state == "" {
		return false
	}
	cookie, err := r.Cookie(stateCookieName)
	if err != nil {
		return false
	}
	value, sig, ok := strings.Cut(cookie.Value, ".")
	if !ok {
		return false
	}
	if !hmac.Equal([]byte(sig), []byte(signState(value, secret))) {
		return false
	}
	return hmac.Equal([]byte(value), []byte(state))
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

func signState(state, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(state))
	return hex.EncodeToString(mac.Sum(nil))
}
