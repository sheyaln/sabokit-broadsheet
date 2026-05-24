package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Notifuse/notifuse/internal/domain"
	http_handler "github.com/Notifuse/notifuse/internal/http"
	pkgmocks "github.com/Notifuse/notifuse/pkg/mocks"
)

// fakeOIDCService is a hand-rolled fake satisfying http.OIDCServiceInterface.
// The interface is too small to justify gomock generation.
type fakeOIDCService struct {
	authorizeURL    string
	authorizeErr    error
	callbackResp    *domain.AuthResponse
	callbackErr     error
	groupMappings   []domain.OIDCGroupMapping
	setMappingsErr  error
	setMappingsArg  []domain.OIDCGroupMapping
	handleCallback  func(ctx context.Context, state, code string) (*domain.AuthResponse, error)
}

func (f *fakeOIDCService) GetAuthorizeURL() (string, error) {
	return f.authorizeURL, f.authorizeErr
}

func (f *fakeOIDCService) HandleCallback(ctx context.Context, state, code string) (*domain.AuthResponse, error) {
	if f.handleCallback != nil {
		return f.handleCallback(ctx, state, code)
	}
	return f.callbackResp, f.callbackErr
}

func (f *fakeOIDCService) GetGroupMappings() []domain.OIDCGroupMapping {
	if f.groupMappings == nil {
		return []domain.OIDCGroupMapping{}
	}
	return f.groupMappings
}

func (f *fakeOIDCService) SetGroupMappings(_ context.Context, mappings []domain.OIDCGroupMapping) error {
	f.setMappingsArg = mappings
	return f.setMappingsErr
}

func setupOIDCHandler(t *testing.T) (*http_handler.OIDCHandler, *fakeOIDCService) {
	ctrl := gomock.NewController(t)
	mockLogger := pkgmocks.NewMockLogger(ctrl)
	mockLogger.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(mockLogger).AnyTimes()
	mockLogger.EXPECT().WithFields(gomock.Any()).Return(mockLogger).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()

	fake := &fakeOIDCService{}
	handler := http_handler.NewOIDCHandler(
		fake,
		func() ([]byte, error) { return []byte("test-jwt-secret-key-for-testing-32bytes"), nil },
		mockLogger,
	)
	return handler, fake
}

func TestOIDCHandler_Authorize(t *testing.T) {
	t.Run("redirects to provider URL", func(t *testing.T) {
		handler, fake := setupOIDCHandler(t)
		fake.authorizeURL = "https://idp.example.com/authorize?state=abc"

		req := httptest.NewRequest(http.MethodGet, "/api/auth/oidc/authorize", nil)
		w := httptest.NewRecorder()
		handler.Authorize(w, req)

		assert.Equal(t, http.StatusTemporaryRedirect, w.Code)
		assert.Equal(t, fake.authorizeURL, w.Header().Get("Location"))
	})

	t.Run("rejects non-GET", func(t *testing.T) {
		handler, _ := setupOIDCHandler(t)
		req := httptest.NewRequest(http.MethodPost, "/api/auth/oidc/authorize", nil)
		w := httptest.NewRecorder()
		handler.Authorize(w, req)
		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})

	t.Run("500 when service returns error", func(t *testing.T) {
		handler, fake := setupOIDCHandler(t)
		fake.authorizeErr = errors.New("rng failure")

		req := httptest.NewRequest(http.MethodGet, "/api/auth/oidc/authorize", nil)
		w := httptest.NewRecorder()
		handler.Authorize(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestOIDCHandler_Callback(t *testing.T) {
	t.Run("happy path redirects to console with token", func(t *testing.T) {
		handler, fake := setupOIDCHandler(t)
		expires := time.Now().Add(24 * time.Hour)
		fake.callbackResp = &domain.AuthResponse{
			Token:     "tok-abc",
			ExpiresAt: expires,
			User:      domain.User{ID: "u-1", Email: "a@b.com"},
		}

		req := httptest.NewRequest(http.MethodGet, "/api/auth/oidc/callback?state=s&code=c", nil)
		w := httptest.NewRecorder()
		handler.Callback(w, req)

		assert.Equal(t, http.StatusTemporaryRedirect, w.Code)
		loc := w.Header().Get("Location")
		assert.True(t, strings.HasPrefix(loc, "/console/auth/oidc/callback?"), "got %s", loc)
		assert.Contains(t, loc, "token=tok-abc")
		assert.Contains(t, loc, "expires_at=")
	})

	t.Run("provider error redirects to signin with message", func(t *testing.T) {
		handler, _ := setupOIDCHandler(t)
		req := httptest.NewRequest(http.MethodGet,
			"/api/auth/oidc/callback?error=access_denied&error_description=user+declined", nil)
		w := httptest.NewRecorder()
		handler.Callback(w, req)

		assert.Equal(t, http.StatusTemporaryRedirect, w.Code)
		loc := w.Header().Get("Location")
		assert.True(t, strings.HasPrefix(loc, "/console/signin?"), "got %s", loc)
		assert.Contains(t, loc, "error=oidc_failed")
		assert.Contains(t, loc, "user+declined")
	})

	t.Run("missing state or code → signin error", func(t *testing.T) {
		handler, _ := setupOIDCHandler(t)

		// Missing both
		req := httptest.NewRequest(http.MethodGet, "/api/auth/oidc/callback", nil)
		w := httptest.NewRecorder()
		handler.Callback(w, req)
		assert.Equal(t, http.StatusTemporaryRedirect, w.Code)
		assert.Contains(t, w.Header().Get("Location"), "/console/signin?")

		// Missing code only
		req = httptest.NewRequest(http.MethodGet, "/api/auth/oidc/callback?state=s", nil)
		w = httptest.NewRecorder()
		handler.Callback(w, req)
		assert.Equal(t, http.StatusTemporaryRedirect, w.Code)
		assert.Contains(t, w.Header().Get("Location"), "/console/signin?")
	})

	t.Run("service callback failure → signin error", func(t *testing.T) {
		handler, fake := setupOIDCHandler(t)
		fake.callbackErr = errors.New("nonce mismatch")

		req := httptest.NewRequest(http.MethodGet, "/api/auth/oidc/callback?state=s&code=c", nil)
		w := httptest.NewRecorder()
		handler.Callback(w, req)

		assert.Equal(t, http.StatusTemporaryRedirect, w.Code)
		loc := w.Header().Get("Location")
		assert.Contains(t, loc, "/console/signin?")
		assert.Contains(t, loc, "SSO+authentication+failed")
	})

	t.Run("rejects non-GET", func(t *testing.T) {
		handler, _ := setupOIDCHandler(t)
		req := httptest.NewRequest(http.MethodPost, "/api/auth/oidc/callback", nil)
		w := httptest.NewRecorder()
		handler.Callback(w, req)
		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})

	t.Run("passes state and code through to service", func(t *testing.T) {
		handler, fake := setupOIDCHandler(t)
		var capturedState, capturedCode string
		fake.handleCallback = func(_ context.Context, state, code string) (*domain.AuthResponse, error) {
			capturedState = state
			capturedCode = code
			return &domain.AuthResponse{Token: "t", ExpiresAt: time.Now().Add(time.Hour)}, nil
		}

		req := httptest.NewRequest(http.MethodGet, "/api/auth/oidc/callback?state=xyz&code=abc", nil)
		w := httptest.NewRecorder()
		handler.Callback(w, req)
		assert.Equal(t, "xyz", capturedState)
		assert.Equal(t, "abc", capturedCode)
	})
}

func TestOIDCHandler_GetGroupMappings(t *testing.T) {
	t.Run("returns mappings as JSON", func(t *testing.T) {
		handler, fake := setupOIDCHandler(t)
		fake.groupMappings = []domain.OIDCGroupMapping{
			{OIDCGroup: "g1", Role: "owner"},
			{OIDCGroup: "g2", Role: "member"},
		}

		req := httptest.NewRequest(http.MethodGet, "/api/oidc.getGroupMappings", nil)
		w := httptest.NewRecorder()
		handler.GetGroupMappings(w, req)

		require.Equal(t, http.StatusOK, w.Code)
		var resp struct {
			Mappings []domain.OIDCGroupMapping `json:"mappings"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, fake.groupMappings, resp.Mappings)
	})

	t.Run("returns empty array when no mappings", func(t *testing.T) {
		handler, _ := setupOIDCHandler(t)
		req := httptest.NewRequest(http.MethodGet, "/api/oidc.getGroupMappings", nil)
		w := httptest.NewRecorder()
		handler.GetGroupMappings(w, req)

		require.Equal(t, http.StatusOK, w.Code)
		var resp struct {
			Mappings []domain.OIDCGroupMapping `json:"mappings"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.NotNil(t, resp.Mappings)
		assert.Empty(t, resp.Mappings)
	})

	t.Run("rejects non-GET", func(t *testing.T) {
		handler, _ := setupOIDCHandler(t)
		req := httptest.NewRequest(http.MethodPost, "/api/oidc.getGroupMappings", nil)
		w := httptest.NewRecorder()
		handler.GetGroupMappings(w, req)
		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})
}

func TestOIDCHandler_SetGroupMappings(t *testing.T) {
	t.Run("saves valid mappings", func(t *testing.T) {
		handler, fake := setupOIDCHandler(t)
		mappings := []domain.OIDCGroupMapping{
			{OIDCGroup: "g1", Role: "owner"},
			{OIDCGroup: "g2", Role: "member", Permissions: domain.UserPermissions{
				domain.PermissionResourceContacts: {Read: true},
			}},
		}
		body, _ := json.Marshal(map[string]interface{}{"mappings": mappings})

		req := httptest.NewRequest(http.MethodPost, "/api/oidc.setGroupMappings", bytes.NewReader(body))
		w := httptest.NewRecorder()
		handler.SetGroupMappings(w, req)

		require.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, mappings, fake.setMappingsArg)
	})

	t.Run("rejects mapping with empty oidc_group", func(t *testing.T) {
		handler, fake := setupOIDCHandler(t)
		body, _ := json.Marshal(map[string]interface{}{
			"mappings": []domain.OIDCGroupMapping{{Role: "owner"}},
		})
		req := httptest.NewRequest(http.MethodPost, "/api/oidc.setGroupMappings", bytes.NewReader(body))
		w := httptest.NewRecorder()
		handler.SetGroupMappings(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "oidc_group is required")
		assert.Nil(t, fake.setMappingsArg, "service must not be called on validation error")
	})

	t.Run("rejects invalid role", func(t *testing.T) {
		handler, _ := setupOIDCHandler(t)
		body, _ := json.Marshal(map[string]interface{}{
			"mappings": []domain.OIDCGroupMapping{{OIDCGroup: "g1", Role: "admin"}},
		})
		req := httptest.NewRequest(http.MethodPost, "/api/oidc.setGroupMappings", bytes.NewReader(body))
		w := httptest.NewRecorder()
		handler.SetGroupMappings(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "role must be")
	})

	t.Run("rejects malformed JSON", func(t *testing.T) {
		handler, _ := setupOIDCHandler(t)
		req := httptest.NewRequest(http.MethodPost, "/api/oidc.setGroupMappings",
			bytes.NewReader([]byte("{not json")))
		w := httptest.NewRecorder()
		handler.SetGroupMappings(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("500 on persistence error", func(t *testing.T) {
		handler, fake := setupOIDCHandler(t)
		fake.setMappingsErr = errors.New("db down")
		body, _ := json.Marshal(map[string]interface{}{
			"mappings": []domain.OIDCGroupMapping{{OIDCGroup: "g1", Role: "owner"}},
		})
		req := httptest.NewRequest(http.MethodPost, "/api/oidc.setGroupMappings", bytes.NewReader(body))
		w := httptest.NewRecorder()
		handler.SetGroupMappings(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("rejects non-POST", func(t *testing.T) {
		handler, _ := setupOIDCHandler(t)
		req := httptest.NewRequest(http.MethodGet, "/api/oidc.setGroupMappings", nil)
		w := httptest.NewRecorder()
		handler.SetGroupMappings(w, req)
		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})

	t.Run("accepts empty mappings list (clears all)", func(t *testing.T) {
		handler, fake := setupOIDCHandler(t)
		body, _ := json.Marshal(map[string]interface{}{
			"mappings": []domain.OIDCGroupMapping{},
		})
		req := httptest.NewRequest(http.MethodPost, "/api/oidc.setGroupMappings", bytes.NewReader(body))
		w := httptest.NewRecorder()
		handler.SetGroupMappings(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Empty(t, fake.setMappingsArg)
	})
}

func TestOIDCMagicCodeGuard(t *testing.T) {
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	cases := []struct {
		name           string
		oidcEnabled    bool
		allowMagicCode bool
		wantStatus     int
		wantCalled     bool
	}{
		{"oidc disabled, magic allowed → pass", false, true, http.StatusOK, true},
		{"oidc disabled, magic disallowed → still pass (oidc off)", false, false, http.StatusOK, true},
		{"oidc enabled, magic allowed → pass", true, true, http.StatusOK, true},
		{"oidc enabled, magic disallowed → 403", true, false, http.StatusForbidden, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			called = false
			guard := http_handler.NewOIDCMagicCodeGuard(tc.oidcEnabled, tc.allowMagicCode)
			req := httptest.NewRequest(http.MethodPost, "/api/auth/signin", nil)
			w := httptest.NewRecorder()
			guard.Guard(next)(w, req)
			assert.Equal(t, tc.wantStatus, w.Code)
			assert.Equal(t, tc.wantCalled, called)
		})
	}
}

func TestOIDCHandler_RegisterRoutes(t *testing.T) {
	handler, fake := setupOIDCHandler(t)
	fake.authorizeURL = "https://idp.example.com/authorize"

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	// Public routes must be reachable without auth.
	req := httptest.NewRequest(http.MethodGet, "/api/auth/oidc/authorize", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	assert.Equal(t, http.StatusTemporaryRedirect, w.Code)

	// Authed mapping routes must demand a token (no auth → 401).
	req = httptest.NewRequest(http.MethodGet, "/api/oidc.getGroupMappings", nil)
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
