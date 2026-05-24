package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/internal/http/middleware"
	"github.com/Notifuse/notifuse/pkg/logger"
)

// OIDCServiceInterface is the slice of the OIDC service the handler depends on.
// Defined here so handler tests can substitute a fake.
type OIDCServiceInterface interface {
	GetAuthorizeURL() (string, error)
	HandleCallback(ctx context.Context, state, code string) (*domain.AuthResponse, error)
	GetGroupMappings() []domain.OIDCGroupMapping
	SetGroupMappings(ctx context.Context, mappings []domain.OIDCGroupMapping) error
}

type OIDCHandler struct {
	oidcService  OIDCServiceInterface
	getJWTSecret func() ([]byte, error)
	logger       logger.Logger
}

func NewOIDCHandler(oidcService OIDCServiceInterface, getJWTSecret func() ([]byte, error), logger logger.Logger) *OIDCHandler {
	return &OIDCHandler{
		oidcService:  oidcService,
		getJWTSecret: getJWTSecret,
		logger:       logger,
	}
}

func (h *OIDCHandler) Authorize(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	authURL, err := h.oidcService.GetAuthorizeURL()
	if err != nil {
		h.logger.WithField("error", err.Error()).Error("Failed to generate OIDC authorize URL")
		WriteJSONError(w, "Failed to initiate SSO login", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

func (h *OIDCHandler) Callback(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if errParam := r.URL.Query().Get("error"); errParam != "" {
		errDesc := r.URL.Query().Get("error_description")
		h.logger.WithField("error", errParam).WithField("description", errDesc).
			Error("OIDC provider returned error")
		redirectWithError(w, r, "SSO provider returned an error: "+errDesc)
		return
	}

	state := r.URL.Query().Get("state")
	code := r.URL.Query().Get("code")

	if state == "" || code == "" {
		redirectWithError(w, r, "Missing state or code parameter")
		return
	}

	ctx := context.Background()
	authResponse, err := h.oidcService.HandleCallback(ctx, state, code)
	if err != nil {
		h.logger.WithField("error", err.Error()).Error("OIDC callback handling failed")
		redirectWithError(w, r, "SSO authentication failed")
		return
	}

	redirectURL := fmt.Sprintf("/console/auth/oidc/callback?token=%s&expires_at=%s",
		url.QueryEscape(authResponse.Token),
		url.QueryEscape(strconv.FormatInt(authResponse.ExpiresAt.Unix(), 10)),
	)

	http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
}

func redirectWithError(w http.ResponseWriter, r *http.Request, message string) {
	redirectURL := fmt.Sprintf("/console/signin?error=oidc_failed&message=%s",
		url.QueryEscape(message),
	)
	http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
}

func (h *OIDCHandler) GetGroupMappings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	mappings := h.oidcService.GetGroupMappings()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"mappings": mappings,
	})
}

func (h *OIDCHandler) SetGroupMappings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Mappings []domain.OIDCGroupMapping `json:"mappings"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	for i, m := range req.Mappings {
		if m.OIDCGroup == "" {
			WriteJSONError(w, fmt.Sprintf("mapping %d: oidc_group is required", i), http.StatusBadRequest)
			return
		}
		if m.Role != "owner" && m.Role != "member" {
			WriteJSONError(w, fmt.Sprintf("mapping %d: role must be 'owner' or 'member'", i), http.StatusBadRequest)
			return
		}
	}

	if err := h.oidcService.SetGroupMappings(r.Context(), req.Mappings); err != nil {
		h.logger.WithField("error", err.Error()).Error("Failed to save OIDC group mappings")
		WriteJSONError(w, "Failed to save group mappings", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"message": "Group mappings saved successfully",
	})
}

func (h *OIDCHandler) RegisterRoutes(mux *http.ServeMux) {
	// Public OIDC flow
	mux.HandleFunc("/api/auth/oidc/authorize", h.Authorize)
	mux.HandleFunc("/api/auth/oidc/callback", h.Callback)

	// Authenticated group mapping management
	authMiddleware := middleware.NewAuthMiddleware(h.getJWTSecret)
	requireAuth := authMiddleware.RequireAuth()
	mux.Handle("/api/oidc.getGroupMappings", requireAuth(http.HandlerFunc(h.GetGroupMappings)))
	mux.Handle("/api/oidc.setGroupMappings", requireAuth(http.HandlerFunc(h.SetGroupMappings)))
}

// OIDCMagicCodeGuard blocks magic code signin when OIDC enforces SSO-only login.
type OIDCMagicCodeGuard struct {
	oidcEnabled    bool
	allowMagicCode bool
}

func NewOIDCMagicCodeGuard(oidcEnabled, allowMagicCode bool) *OIDCMagicCodeGuard {
	return &OIDCMagicCodeGuard{
		oidcEnabled:    oidcEnabled,
		allowMagicCode: allowMagicCode,
	}
}

func (g *OIDCMagicCodeGuard) Guard(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if g.oidcEnabled && !g.allowMagicCode {
			WriteJSONError(w, "Magic code login is disabled when SSO is enforced", http.StatusForbidden)
			return
		}
		next(w, r)
	}
}

