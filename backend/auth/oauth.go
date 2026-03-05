package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"

	"github.com/simu/drum-trainer/backend/config"
)

// UserInfo holds the normalised user profile returned by an OAuth provider.
type UserInfo struct {
	ProviderID string
	Email      string
	Name       string
	AvatarURL  string
}

// NewConfig returns an oauth2.Config for the given provider name.
func NewConfig(provider string, cfg config.Config) (*oauth2.Config, error) {
	switch provider {
	case "google":
		return &oauth2.Config{
			ClientID:     cfg.GoogleClientID,
			ClientSecret: cfg.GoogleClientSecret,
			RedirectURL:  cfg.BaseURL + "/api/v1/auth/google/callback",
			Scopes:       []string{"openid", "email", "profile"},
			Endpoint:     google.Endpoint,
		}, nil
	case "github":
		return &oauth2.Config{
			ClientID:     cfg.GitHubClientID,
			ClientSecret: cfg.GitHubClientSecret,
			RedirectURL:  cfg.BaseURL + "/api/v1/auth/github/callback",
			Scopes:       []string{"user:email"},
			Endpoint:     github.Endpoint,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}
}

// FetchUserInfo retrieves the user's profile from the OAuth provider.
func FetchUserInfo(ctx context.Context, provider string, token *oauth2.Token, conf *oauth2.Config) (*UserInfo, error) {
	client := conf.Client(ctx, token)
	switch provider {
	case "google":
		return fetchGoogleUser(client)
	case "github":
		return fetchGitHubUser(ctx, client)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}
}

// --- Google ---

type googleUser struct {
	ID      string `json:"id"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}

func fetchGoogleUser(client *http.Client) (*UserInfo, error) {
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, fmt.Errorf("google userinfo request: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var u googleUser
	if err := json.Unmarshal(body, &u); err != nil {
		return nil, fmt.Errorf("google userinfo decode: %w", err)
	}
	return &UserInfo{
		ProviderID: u.ID,
		Email:      u.Email,
		Name:       u.Name,
		AvatarURL:  u.Picture,
	}, nil
}

// --- GitHub ---

type githubUser struct {
	ID        int    `json:"id"`
	Login     string `json:"login"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
}

type githubEmail struct {
	Email    string `json:"email"`
	Primary  bool   `json:"primary"`
	Verified bool   `json:"verified"`
}

func fetchGitHubUser(ctx context.Context, client *http.Client) (*UserInfo, error) {
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		return nil, fmt.Errorf("github user request: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var u githubUser
	if err := json.Unmarshal(body, &u); err != nil {
		return nil, fmt.Errorf("github user decode: %w", err)
	}

	name := u.Name
	if name == "" {
		name = u.Login
	}

	email := u.Email
	if email == "" {
		// GitHub may hide the primary email in the profile; fetch it separately.
		email, _ = fetchGitHubPrimaryEmail(client)
	}

	return &UserInfo{
		ProviderID: fmt.Sprintf("%d", u.ID),
		Email:      email,
		Name:       name,
		AvatarURL:  u.AvatarURL,
	}, nil
}

func fetchGitHubPrimaryEmail(client *http.Client) (string, error) {
	resp, err := client.Get("https://api.github.com/user/emails")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	var emails []githubEmail
	if err := json.Unmarshal(body, &emails); err != nil {
		return "", err
	}
	for _, e := range emails {
		if e.Primary && e.Verified {
			return e.Email, nil
		}
	}
	return "", nil
}
