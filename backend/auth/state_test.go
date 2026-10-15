package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// requestWithStateCookie issues a state via GenerateState and returns a
// request carrying the resulting cookie, plus the state value.
func requestWithStateCookie(t *testing.T, secret string) (*http.Request, string) {
	t.Helper()
	rec := httptest.NewRecorder()
	state, err := GenerateState(rec, secret, true)
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodGet, "/callback", nil)
	for _, c := range rec.Result().Cookies() {
		req.AddCookie(c)
	}
	return req, state
}

func TestVerifyStateAcceptsMatchingSignedState(t *testing.T) {
	req, state := requestWithStateCookie(t, "secret")
	if !VerifyState(req, state, "secret") {
		t.Fatal("expected valid state to verify")
	}
}

func TestVerifyStateRejectsWrongState(t *testing.T) {
	req, _ := requestWithStateCookie(t, "secret")
	if VerifyState(req, "deadbeef", "secret") {
		t.Fatal("expected mismatched state to be rejected")
	}
}

func TestVerifyStateRejectsWrongSecret(t *testing.T) {
	req, state := requestWithStateCookie(t, "secret")
	if VerifyState(req, state, "other-secret") {
		t.Fatal("expected state signed with another secret to be rejected")
	}
}

func TestVerifyStateRejectsUnsignedCookie(t *testing.T) {
	// A planted cookie whose value simply equals the query parameter.
	req := httptest.NewRequest(http.MethodGet, "/callback", nil)
	req.AddCookie(&http.Cookie{Name: stateCookieName, Value: "abc"})
	if VerifyState(req, "abc", "secret") {
		t.Fatal("expected unsigned cookie to be rejected")
	}
}
