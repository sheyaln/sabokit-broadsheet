package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/google/uuid"
	"golang.org/x/oauth2"

	"github.com/sheyaln/sabokit-broadside/config"
	"github.com/sheyaln/sabokit-broadside/internal/domain"
	"github.com/sheyaln/sabokit-broadside/pkg/cache"
	"github.com/sheyaln/sabokit-broadside/pkg/logger"
)

const (
	oidcStateTTL   = 10 * time.Minute
	sessionExpiry  = 30 * 24 * time.Hour // 30 days, matches user service
	oidcCacheClean = 1 * time.Minute
)

type OIDCService struct {
	provider      *oidc.Provider
	oauth2Config  *oauth2.Config
	verifier      *oidc.IDTokenVerifier
	stateCache    cache.Cache
	userRepo      domain.UserRepository
	workspaceRepo domain.WorkspaceRepository
	authService   domain.AuthService
	settingService *SettingService
	config        *config.OIDCConfig
	groupMappings []domain.OIDCGroupMapping
	logger        logger.Logger
	secretKey     string
}

type OIDCServiceConfig struct {
	Config         *config.OIDCConfig
	APIEndpoint    string
	UserRepo       domain.UserRepository
	WorkspaceRepo  domain.WorkspaceRepository
	AuthService    domain.AuthService
	SettingService *SettingService
	SecretKey      string
	Logger         logger.Logger
}

func NewOIDCService(ctx context.Context, cfg OIDCServiceConfig) (*OIDCService, error) {
	if !cfg.Config.Enabled {
		return nil, fmt.Errorf("OIDC is not enabled")
	}

	if cfg.Config.IssuerURL == "" || cfg.Config.ClientID == "" || cfg.Config.ClientSecret == "" {
		return nil, fmt.Errorf("OIDC requires issuer_url, client_id, and client_secret")
	}

	provider, err := oidc.NewProvider(ctx, cfg.Config.IssuerURL)
	if err != nil {
		return nil, fmt.Errorf("OIDC discovery failed for %s: %w", cfg.Config.IssuerURL, err)
	}

	callbackURL := cfg.APIEndpoint + "/api/auth/oidc/callback"

	oauth2Config := &oauth2.Config{
		ClientID:     cfg.Config.ClientID,
		ClientSecret: cfg.Config.ClientSecret,
		Endpoint:     provider.Endpoint(),
		RedirectURL:  callbackURL,
		Scopes:       []string{oidc.ScopeOpenID, "email", "profile", "groups"},
	}

	verifier := provider.Verifier(&oidc.Config{ClientID: cfg.Config.ClientID})

	svc := &OIDCService{
		provider:       provider,
		oauth2Config:   oauth2Config,
		verifier:       verifier,
		stateCache:     cache.NewInMemoryCache(oidcCacheClean),
		userRepo:       cfg.UserRepo,
		workspaceRepo:  cfg.WorkspaceRepo,
		authService:    cfg.AuthService,
		settingService: cfg.SettingService,
		config:         cfg.Config,
		logger:         cfg.Logger,
		secretKey:      cfg.SecretKey,
	}

	if err := svc.ReloadGroupMappings(ctx); err != nil {
		svc.logger.WithField("error", err.Error()).Warn("Failed to load OIDC group mappings, continuing without")
	}

	return svc, nil
}

func (s *OIDCService) GetAuthorizeURL() (string, error) {
	stateBytes := make([]byte, 32)
	if _, err := rand.Read(stateBytes); err != nil {
		return "", fmt.Errorf("failed to generate state: %w", err)
	}
	state := hex.EncodeToString(stateBytes)

	nonceBytes := make([]byte, 32)
	if _, err := rand.Read(nonceBytes); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}
	nonce := hex.EncodeToString(nonceBytes)

	oidcState := &domain.OIDCState{
		State:     state,
		Nonce:     nonce,
		CreatedAt: time.Now(),
	}
	s.stateCache.Set(state, oidcState, oidcStateTTL)

	url := s.oauth2Config.AuthCodeURL(state, oidc.Nonce(nonce))
	return url, nil
}

func (s *OIDCService) HandleCallback(ctx context.Context, stateParam, code string) (*domain.AuthResponse, error) {
	cached, found := s.stateCache.Get(stateParam)
	if !found {
		return nil, fmt.Errorf("invalid or expired OIDC state")
	}
	s.stateCache.Delete(stateParam)

	oidcState, ok := cached.(*domain.OIDCState)
	if !ok {
		return nil, fmt.Errorf("corrupted OIDC state")
	}

	oauth2Token, err := s.oauth2Config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange authorization code: %w", err)
	}

	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		return nil, fmt.Errorf("no id_token in token response")
	}

	idToken, err := s.verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return nil, fmt.Errorf("failed to verify ID token: %w", err)
	}

	if idToken.Nonce != oidcState.Nonce {
		return nil, fmt.Errorf("nonce mismatch")
	}

	var claims struct {
		Email             string   `json:"email"`
		Name              string   `json:"name"`
		PreferredUsername string   `json:"preferred_username"`
		Groups            []string `json:"groups"`
	}
	if err := idToken.Claims(&claims); err != nil {
		return nil, fmt.Errorf("failed to extract ID token claims: %w", err)
	}

	if claims.Email == "" {
		return nil, fmt.Errorf("email claim is missing from ID token")
	}

	// Some IdPs use a custom claim name for groups instead of "groups"
	if len(claims.Groups) == 0 && s.config.GroupsClaim != "" && s.config.GroupsClaim != "groups" {
		var rawClaims map[string]interface{}
		if err := idToken.Claims(&rawClaims); err == nil {
			if groupsRaw, ok := rawClaims[s.config.GroupsClaim]; ok {
				if groupsList, ok := groupsRaw.([]interface{}); ok {
					for _, g := range groupsList {
						if gs, ok := g.(string); ok {
							claims.Groups = append(claims.Groups, gs)
						}
					}
				}
			}
		}
	}

	name := claims.Name
	if name == "" {
		name = claims.PreferredUsername
	}

	user, isNew, err := s.findOrCreateUser(ctx, claims.Email, name)
	if err != nil {
		return nil, fmt.Errorf("user provisioning failed: %w", err)
	}

	if err := s.syncWorkspaceAccess(ctx, user, claims.Groups, isNew); err != nil {
		s.logger.WithField("error", err.Error()).WithField("user_id", user.ID).
			Error("Failed to sync OIDC workspace access")
	}

	expiresAt := time.Now().Add(sessionExpiry)
	session := &domain.Session{
		ID:        generateUUID(),
		UserID:    user.ID,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
	}
	if err := s.userRepo.CreateSession(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	token := s.authService.GenerateUserAuthToken(user, session.ID, expiresAt)

	return &domain.AuthResponse{
		Token:     token,
		User:      *user,
		ExpiresAt: expiresAt,
	}, nil
}

func (s *OIDCService) findOrCreateUser(ctx context.Context, email, name string) (*domain.User, bool, error) {
	user, err := s.userRepo.GetUserByEmail(ctx, email)
	if err == nil {
		return user, false, nil
	}

	if _, ok := err.(*domain.ErrUserNotFound); !ok {
		return nil, false, fmt.Errorf("failed to look up user: %w", err)
	}

	if !s.config.AutoProvision {
		return nil, false, fmt.Errorf("user %s not provisioned and auto-provisioning is disabled", email)
	}

	newUser := &domain.User{
		ID:        generateUUID(),
		Type:      domain.UserTypeUser,
		Email:     email,
		Name:      name,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.userRepo.CreateUser(ctx, newUser); err != nil {
		return nil, false, fmt.Errorf("failed to create user: %w", err)
	}

	s.logger.WithField("email", email).WithField("user_id", newUser.ID).Info("OIDC auto-provisioned new user")
	return newUser, true, nil
}

// syncWorkspaceAccess ensures OIDC users have workspace access.
// If group mappings match, those permissions are applied. Otherwise, newly
// provisioned users get member access with full permissions to all workspaces.
func (s *OIDCService) syncWorkspaceAccess(ctx context.Context, user *domain.User, groups []string, isNew bool) error {
	role, permissions, matched := s.resolveGroupPermissions(groups)

	// For new users with no matching group mappings, grant default access
	if !matched && isNew {
		role = "member"
		permissions = domain.FullPermissions
	}

	if !matched && !isNew {
		return nil
	}

	return s.applyWorkspaceAccess(ctx, user, role, permissions)
}

func (s *OIDCService) resolveGroupPermissions(groups []string) (string, domain.UserPermissions, bool) {
	if len(groups) == 0 || len(s.groupMappings) == 0 {
		return "", nil, false
	}

	groupSet := make(map[string]bool)
	for _, g := range groups {
		groupSet[g] = true
	}

	role := "member"
	permissions := make(domain.UserPermissions)
	anyMatched := false

	for _, mapping := range s.groupMappings {
		if !groupSet[mapping.OIDCGroup] {
			continue
		}
		anyMatched = true
		if mapping.Role == "owner" {
			role = "owner"
		}
		for resource, perms := range mapping.Permissions {
			existing := permissions[resource]
			if perms.Read {
				existing.Read = true
			}
			if perms.Write {
				existing.Write = true
			}
			permissions[resource] = existing
		}
	}

	return role, permissions, anyMatched
}

func (s *OIDCService) applyWorkspaceAccess(ctx context.Context, user *domain.User, role string, permissions domain.UserPermissions) error {
	workspaces, err := s.workspaceRepo.List(ctx)
	if err != nil {
		return fmt.Errorf("failed to list workspaces: %w", err)
	}

	for _, ws := range workspaces {
		existing, err := s.workspaceRepo.GetUserWorkspace(ctx, user.ID, ws.ID)
		if err != nil {
			uw := &domain.UserWorkspace{
				UserID:      user.ID,
				WorkspaceID: ws.ID,
				Role:        role,
				Permissions: permissions,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}
			if addErr := s.workspaceRepo.AddUserToWorkspace(ctx, uw); addErr != nil {
				s.logger.WithField("error", addErr.Error()).
					WithField("workspace_id", ws.ID).
					WithField("user_id", user.ID).
					Error("Failed to add user to workspace via OIDC")
			}
			continue
		}

		existing.Role = role
		existing.Permissions = permissions
		existing.UpdatedAt = time.Now()
		if updateErr := s.workspaceRepo.UpdateUserWorkspacePermissions(ctx, existing); updateErr != nil {
			s.logger.WithField("error", updateErr.Error()).
				WithField("workspace_id", ws.ID).
				WithField("user_id", user.ID).
				Error("Failed to update workspace permissions via OIDC")
		}
	}

	return nil
}

func (s *OIDCService) ReloadGroupMappings(ctx context.Context) error {
	value, err := s.settingService.GetSetting(ctx, "oidc_group_mappings")
	if err != nil {
		s.groupMappings = nil
		return nil
	}

	if value == "" {
		s.groupMappings = nil
		return nil
	}

	var mappings []domain.OIDCGroupMapping
	if err := json.Unmarshal([]byte(value), &mappings); err != nil {
		return fmt.Errorf("failed to parse oidc_group_mappings: %w", err)
	}

	s.groupMappings = mappings
	s.logger.WithField("count", len(mappings)).Info("Loaded OIDC group mappings")
	return nil
}

func (s *OIDCService) GetGroupMappings() []domain.OIDCGroupMapping {
	if s.groupMappings == nil {
		return []domain.OIDCGroupMapping{}
	}
	return s.groupMappings
}

func (s *OIDCService) SetGroupMappings(ctx context.Context, mappings []domain.OIDCGroupMapping) error {
	data, err := json.Marshal(mappings)
	if err != nil {
		return fmt.Errorf("failed to marshal group mappings: %w", err)
	}

	if err := s.settingService.SetSetting(ctx, "oidc_group_mappings", string(data)); err != nil {
		return fmt.Errorf("failed to save group mappings: %w", err)
	}

	s.groupMappings = mappings
	s.logger.WithField("count", len(mappings)).Info("Updated OIDC group mappings")
	return nil
}

func (s *OIDCService) IsEnabled() bool {
	return s.config.Enabled
}

func (s *OIDCService) AllowMagicCode() bool {
	return s.config.AllowMagicCode
}

func (s *OIDCService) Stop() {
	if s.stateCache != nil {
		s.stateCache.Stop()
	}
}

func generateUUID() string {
	return uuid.New().String()
}
