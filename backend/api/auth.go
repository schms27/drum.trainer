package api

import (
	"database/sql"
	"errors"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	authpkg "github.com/simu/drum-trainer/backend/auth"
	"github.com/simu/drum-trainer/backend/config"
	dbpkg "github.com/simu/drum-trainer/backend/db"
)

const sessionCookieName = "drum_session"
const sessionMaxAge = 30 * 24 * 60 * 60 // 30 days in seconds

type authHandler struct {
	db  *sql.DB
	cfg config.Config
}

// GET /api/v1/auth/{provider}/login
func (h *authHandler) login(w http.ResponseWriter, r *http.Request) {
	provider := chi.URLParam(r, "provider")
	conf, err := authpkg.NewConfig(provider, h.cfg)
	if err != nil {
		writeError(w, http.StatusBadRequest, "unsupported provider")
		return
	}

	state, err := authpkg.GenerateState(w, h.cfg.DevMode)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to generate state")
		return
	}

	http.Redirect(w, r, conf.AuthCodeURL(state), http.StatusTemporaryRedirect)
}

// GET /api/v1/auth/{provider}/callback
func (h *authHandler) callback(w http.ResponseWriter, r *http.Request) {
	provider := chi.URLParam(r, "provider")

	if !authpkg.VerifyState(r, r.URL.Query().Get("state")) {
		http.Redirect(w, r, "/?auth_error=state_mismatch", http.StatusTemporaryRedirect)
		return
	}
	authpkg.ClearState(w)

	conf, err := authpkg.NewConfig(provider, h.cfg)
	if err != nil {
		http.Redirect(w, r, "/?auth_error=unsupported_provider", http.StatusTemporaryRedirect)
		return
	}

	oauthToken, err := conf.Exchange(r.Context(), r.URL.Query().Get("code"))
	if err != nil {
		log.Printf("oauth exchange error (%s): %v", provider, err)
		http.Redirect(w, r, "/?auth_error=oauth_exchange", http.StatusTemporaryRedirect)
		return
	}

	info, err := authpkg.FetchUserInfo(r.Context(), provider, oauthToken, conf)
	if err != nil {
		log.Printf("fetch userinfo error (%s): %v", provider, err)
		http.Redirect(w, r, "/?auth_error=userinfo", http.StatusTemporaryRedirect)
		return
	}

	user, err := dbpkg.UpsertUser(h.db, provider, info.ProviderID, info.Email, info.Name, info.AvatarURL)
	if err != nil {
		log.Printf("upsert user error: %v", err)
		http.Redirect(w, r, "/?auth_error=db", http.StatusTemporaryRedirect)
		return
	}

	sessionToken, err := authpkg.GenerateToken()
	if err != nil {
		http.Redirect(w, r, "/?auth_error=session", http.StatusTemporaryRedirect)
		return
	}

	if err := dbpkg.CreateSession(h.db, sessionToken, user.ID); err != nil {
		log.Printf("create session error: %v", err)
		http.Redirect(w, r, "/?auth_error=session", http.StatusTemporaryRedirect)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    sessionToken,
		Path:     "/",
		MaxAge:   sessionMaxAge,
		HttpOnly: true,
		Secure:   !h.cfg.DevMode,
		SameSite: http.SameSiteLaxMode,
	})

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

// POST /api/v1/auth/logout
func (h *authHandler) logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(sessionCookieName)
	if err == nil {
		// Best-effort deletion — ignore error (already expired tokens are harmless).
		dbpkg.DeleteSession(h.db, cookie.Value)
	}
	http.SetCookie(w, &http.Cookie{
		Name:   sessionCookieName,
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
	w.WriteHeader(http.StatusNoContent)
}

// GET /api/v1/auth/me — returns the current user, or 401 if not authenticated.
// This endpoint is intentionally outside the RequireAuth middleware so that the
// frontend can call it to check login state without getting a 401 from middleware.
func (h *authHandler) me(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}
	user, err := dbpkg.ValidateSession(h.db, cookie.Value)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusUnauthorized, "not authenticated")
		} else {
			writeError(w, http.StatusInternalServerError, "session error")
		}
		return
	}
	writeJSON(w, http.StatusOK, user)
}
