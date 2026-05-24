package service

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Notifuse/notifuse/config"
	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/internal/domain/mocks"
	"github.com/Notifuse/notifuse/pkg/cache"
	pkgmocks "github.com/Notifuse/notifuse/pkg/mocks"
)

func newTestOIDCService(t *testing.T) (
	*OIDCService,
	*mocks.MockUserRepository,
	*mocks.MockWorkspaceRepository,
	*mocks.MockSettingRepository,
	*gomock.Controller,
) {
	ctrl := gomock.NewController(t)
	userRepo := mocks.NewMockUserRepository(ctrl)
	wsRepo := mocks.NewMockWorkspaceRepository(ctrl)
	settingRepo := mocks.NewMockSettingRepository(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)

	mockLogger.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(mockLogger).AnyTimes()
	mockLogger.EXPECT().WithFields(gomock.Any()).Return(mockLogger).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()

	svc := &OIDCService{
		stateCache:     cache.NewInMemoryCache(1 * time.Minute),
		userRepo:       userRepo,
		workspaceRepo:  wsRepo,
		settingService: NewSettingService(settingRepo),
		config: &config.OIDCConfig{
			Enabled:        true,
			AutoProvision:  true,
			AllowMagicCode: true,
		},
		logger: mockLogger,
	}
	t.Cleanup(svc.Stop)
	return svc, userRepo, wsRepo, settingRepo, ctrl
}

func TestOIDCService_ResolveGroupPermissions(t *testing.T) {
	svc, _, _, _, ctrl := newTestOIDCService(t)
	defer ctrl.Finish()

	svc.groupMappings = []domain.OIDCGroupMapping{
		{
			OIDCGroup: "notifuse-admins",
			Role:      "owner",
			Permissions: domain.UserPermissions{
				domain.PermissionResourceWorkspace: {Read: true, Write: true},
			},
		},
		{
			OIDCGroup: "notifuse-marketers",
			Role:      "member",
			Permissions: domain.UserPermissions{
				domain.PermissionResourceContacts:  {Read: true, Write: true},
				domain.PermissionResourceTemplates: {Read: true},
			},
		},
	}

	t.Run("no groups → no match", func(t *testing.T) {
		role, perms, matched := svc.resolveGroupPermissions(nil)
		assert.False(t, matched)
		assert.Empty(t, role)
		assert.Nil(t, perms)
	})

	t.Run("unmatched group → no match", func(t *testing.T) {
		_, _, matched := svc.resolveGroupPermissions([]string{"random"})
		assert.False(t, matched)
	})

	t.Run("member group resolves to member with merged perms", func(t *testing.T) {
		role, perms, matched := svc.resolveGroupPermissions([]string{"notifuse-marketers"})
		require.True(t, matched)
		assert.Equal(t, "member", role)
		assert.True(t, perms[domain.PermissionResourceContacts].Read)
		assert.True(t, perms[domain.PermissionResourceContacts].Write)
		assert.True(t, perms[domain.PermissionResourceTemplates].Read)
		assert.False(t, perms[domain.PermissionResourceTemplates].Write)
	})

	t.Run("owner mapping promotes role to owner", func(t *testing.T) {
		role, _, matched := svc.resolveGroupPermissions([]string{"notifuse-admins"})
		require.True(t, matched)
		assert.Equal(t, "owner", role)
	})

	t.Run("multiple groups merge: owner wins, perms union", func(t *testing.T) {
		role, perms, matched := svc.resolveGroupPermissions([]string{"notifuse-marketers", "notifuse-admins"})
		require.True(t, matched)
		assert.Equal(t, "owner", role)
		assert.True(t, perms[domain.PermissionResourceContacts].Write)
		assert.True(t, perms[domain.PermissionResourceWorkspace].Write)
	})

	t.Run("empty mappings → no match even with groups", func(t *testing.T) {
		svc.groupMappings = nil
		_, _, matched := svc.resolveGroupPermissions([]string{"any"})
		assert.False(t, matched)
	})
}

func TestOIDCService_FindOrCreateUser(t *testing.T) {
	t.Run("existing user returned as-is", func(t *testing.T) {
		svc, userRepo, _, _, ctrl := newTestOIDCService(t)
		defer ctrl.Finish()

		existing := &domain.User{ID: "u-1", Email: "a@b.com", Name: "Alice"}
		userRepo.EXPECT().GetUserByEmail(gomock.Any(), "a@b.com").Return(existing, nil)

		user, isNew, err := svc.findOrCreateUser(context.Background(), "a@b.com", "Alice")
		require.NoError(t, err)
		assert.False(t, isNew)
		assert.Equal(t, existing, user)
	})

	t.Run("new user is provisioned when auto-provisioning enabled", func(t *testing.T) {
		svc, userRepo, _, _, ctrl := newTestOIDCService(t)
		defer ctrl.Finish()

		userRepo.EXPECT().GetUserByEmail(gomock.Any(), "new@b.com").
			Return(nil, &domain.ErrUserNotFound{Message: "not found"})
		userRepo.EXPECT().CreateUser(gomock.Any(), gomock.AssignableToTypeOf(&domain.User{})).
			DoAndReturn(func(_ context.Context, u *domain.User) error {
				assert.Equal(t, "new@b.com", u.Email)
				assert.Equal(t, "New User", u.Name)
				assert.Equal(t, domain.UserTypeUser, u.Type)
				assert.NotEmpty(t, u.ID)
				return nil
			})

		user, isNew, err := svc.findOrCreateUser(context.Background(), "new@b.com", "New User")
		require.NoError(t, err)
		assert.True(t, isNew)
		assert.Equal(t, "new@b.com", user.Email)
	})

	t.Run("new user rejected when auto-provisioning disabled", func(t *testing.T) {
		svc, userRepo, _, _, ctrl := newTestOIDCService(t)
		defer ctrl.Finish()
		svc.config.AutoProvision = false

		userRepo.EXPECT().GetUserByEmail(gomock.Any(), "new@b.com").
			Return(nil, &domain.ErrUserNotFound{Message: "not found"})

		_, _, err := svc.findOrCreateUser(context.Background(), "new@b.com", "New User")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "auto-provisioning is disabled")
	})

	t.Run("unexpected lookup error propagates", func(t *testing.T) {
		svc, userRepo, _, _, ctrl := newTestOIDCService(t)
		defer ctrl.Finish()

		userRepo.EXPECT().GetUserByEmail(gomock.Any(), "x@b.com").
			Return(nil, errors.New("db down"))

		_, _, err := svc.findOrCreateUser(context.Background(), "x@b.com", "X")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "db down")
	})

	t.Run("create failure propagates", func(t *testing.T) {
		svc, userRepo, _, _, ctrl := newTestOIDCService(t)
		defer ctrl.Finish()

		userRepo.EXPECT().GetUserByEmail(gomock.Any(), "new@b.com").
			Return(nil, &domain.ErrUserNotFound{Message: "not found"})
		userRepo.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(errors.New("insert failed"))

		_, _, err := svc.findOrCreateUser(context.Background(), "new@b.com", "New")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "insert failed")
	})
}

func TestOIDCService_ApplyWorkspaceAccess(t *testing.T) {
	t.Run("creates user-workspace rows for workspaces user doesn't belong to", func(t *testing.T) {
		svc, _, wsRepo, _, ctrl := newTestOIDCService(t)
		defer ctrl.Finish()

		user := &domain.User{ID: "u-1"}
		workspaces := []*domain.Workspace{{ID: "ws-1"}, {ID: "ws-2"}}
		wsRepo.EXPECT().List(gomock.Any()).Return(workspaces, nil)

		// Both workspaces: user not yet a member → AddUserToWorkspace called for each
		wsRepo.EXPECT().GetUserWorkspace(gomock.Any(), "u-1", "ws-1").
			Return(nil, errors.New("not found"))
		wsRepo.EXPECT().AddUserToWorkspace(gomock.Any(), gomock.AssignableToTypeOf(&domain.UserWorkspace{})).
			DoAndReturn(func(_ context.Context, uw *domain.UserWorkspace) error {
				assert.Equal(t, "u-1", uw.UserID)
				assert.Equal(t, "ws-1", uw.WorkspaceID)
				assert.Equal(t, "owner", uw.Role)
				return nil
			})

		wsRepo.EXPECT().GetUserWorkspace(gomock.Any(), "u-1", "ws-2").
			Return(nil, errors.New("not found"))
		wsRepo.EXPECT().AddUserToWorkspace(gomock.Any(), gomock.Any()).Return(nil)

		err := svc.applyWorkspaceAccess(context.Background(), user, "owner", domain.FullPermissions)
		require.NoError(t, err)
	})

	t.Run("updates permissions when user already in workspace", func(t *testing.T) {
		svc, _, wsRepo, _, ctrl := newTestOIDCService(t)
		defer ctrl.Finish()

		user := &domain.User{ID: "u-1"}
		wsRepo.EXPECT().List(gomock.Any()).Return([]*domain.Workspace{{ID: "ws-1"}}, nil)
		existing := &domain.UserWorkspace{UserID: "u-1", WorkspaceID: "ws-1", Role: "member"}
		wsRepo.EXPECT().GetUserWorkspace(gomock.Any(), "u-1", "ws-1").Return(existing, nil)
		wsRepo.EXPECT().UpdateUserWorkspacePermissions(gomock.Any(), gomock.AssignableToTypeOf(&domain.UserWorkspace{})).
			DoAndReturn(func(_ context.Context, uw *domain.UserWorkspace) error {
				assert.Equal(t, "owner", uw.Role)
				return nil
			})

		err := svc.applyWorkspaceAccess(context.Background(), user, "owner", domain.FullPermissions)
		require.NoError(t, err)
	})

	t.Run("workspace list failure propagates", func(t *testing.T) {
		svc, _, wsRepo, _, ctrl := newTestOIDCService(t)
		defer ctrl.Finish()

		wsRepo.EXPECT().List(gomock.Any()).Return(nil, errors.New("db error"))

		err := svc.applyWorkspaceAccess(context.Background(), &domain.User{ID: "u-1"}, "owner", nil)
		require.Error(t, err)
	})

	t.Run("per-workspace add/update errors are logged but don't abort", func(t *testing.T) {
		svc, _, wsRepo, _, ctrl := newTestOIDCService(t)
		defer ctrl.Finish()

		user := &domain.User{ID: "u-1"}
		wsRepo.EXPECT().List(gomock.Any()).Return([]*domain.Workspace{{ID: "ws-1"}, {ID: "ws-2"}}, nil)

		wsRepo.EXPECT().GetUserWorkspace(gomock.Any(), "u-1", "ws-1").Return(nil, errors.New("not found"))
		wsRepo.EXPECT().AddUserToWorkspace(gomock.Any(), gomock.Any()).Return(errors.New("add failed"))

		wsRepo.EXPECT().GetUserWorkspace(gomock.Any(), "u-1", "ws-2").Return(&domain.UserWorkspace{UserID: "u-1", WorkspaceID: "ws-2"}, nil)
		wsRepo.EXPECT().UpdateUserWorkspacePermissions(gomock.Any(), gomock.Any()).Return(errors.New("update failed"))

		// Function still returns nil — failures are logged
		err := svc.applyWorkspaceAccess(context.Background(), user, "owner", domain.FullPermissions)
		require.NoError(t, err)
	})
}

func TestOIDCService_SyncWorkspaceAccess(t *testing.T) {
	t.Run("matched group → applies mapped permissions", func(t *testing.T) {
		svc, _, wsRepo, _, ctrl := newTestOIDCService(t)
		defer ctrl.Finish()

		svc.groupMappings = []domain.OIDCGroupMapping{
			{OIDCGroup: "g1", Role: "member", Permissions: domain.UserPermissions{
				domain.PermissionResourceContacts: {Read: true},
			}},
		}
		user := &domain.User{ID: "u-1"}
		wsRepo.EXPECT().List(gomock.Any()).Return([]*domain.Workspace{{ID: "ws-1"}}, nil)
		wsRepo.EXPECT().GetUserWorkspace(gomock.Any(), "u-1", "ws-1").Return(nil, errors.New("not found"))
		wsRepo.EXPECT().AddUserToWorkspace(gomock.Any(), gomock.AssignableToTypeOf(&domain.UserWorkspace{})).
			DoAndReturn(func(_ context.Context, uw *domain.UserWorkspace) error {
				assert.Equal(t, "member", uw.Role)
				assert.True(t, uw.Permissions[domain.PermissionResourceContacts].Read)
				return nil
			})

		require.NoError(t, svc.syncWorkspaceAccess(context.Background(), user, []string{"g1"}, false))
	})

	t.Run("new user with no group match → full default access", func(t *testing.T) {
		svc, _, wsRepo, _, ctrl := newTestOIDCService(t)
		defer ctrl.Finish()

		user := &domain.User{ID: "u-1"}
		wsRepo.EXPECT().List(gomock.Any()).Return([]*domain.Workspace{{ID: "ws-1"}}, nil)
		wsRepo.EXPECT().GetUserWorkspace(gomock.Any(), "u-1", "ws-1").Return(nil, errors.New("not found"))
		wsRepo.EXPECT().AddUserToWorkspace(gomock.Any(), gomock.AssignableToTypeOf(&domain.UserWorkspace{})).
			DoAndReturn(func(_ context.Context, uw *domain.UserWorkspace) error {
				assert.Equal(t, "member", uw.Role)
				assert.Equal(t, domain.FullPermissions, uw.Permissions)
				return nil
			})

		require.NoError(t, svc.syncWorkspaceAccess(context.Background(), user, nil, true))
	})

	t.Run("existing user with no group match → no-op", func(t *testing.T) {
		svc, _, _, _, ctrl := newTestOIDCService(t)
		defer ctrl.Finish()

		// No mock expectations: workspaceRepo must not be touched.
		require.NoError(t, svc.syncWorkspaceAccess(context.Background(), &domain.User{ID: "u-1"}, nil, false))
	})
}

func TestOIDCService_GroupMappingsPersistence(t *testing.T) {
	t.Run("ReloadGroupMappings parses valid JSON", func(t *testing.T) {
		svc, _, _, settingRepo, ctrl := newTestOIDCService(t)
		defer ctrl.Finish()

		mappings := []domain.OIDCGroupMapping{
			{OIDCGroup: "g1", Role: "owner"},
		}
		data, _ := json.Marshal(mappings)
		settingRepo.EXPECT().Get(gomock.Any(), "oidc_group_mappings").
			Return(&domain.Setting{Key: "oidc_group_mappings", Value: string(data)}, nil)

		require.NoError(t, svc.ReloadGroupMappings(context.Background()))
		assert.Len(t, svc.GetGroupMappings(), 1)
		assert.Equal(t, "g1", svc.GetGroupMappings()[0].OIDCGroup)
	})

	t.Run("ReloadGroupMappings handles missing setting", func(t *testing.T) {
		svc, _, _, settingRepo, ctrl := newTestOIDCService(t)
		defer ctrl.Finish()

		svc.groupMappings = []domain.OIDCGroupMapping{{OIDCGroup: "old"}}
		settingRepo.EXPECT().Get(gomock.Any(), "oidc_group_mappings").
			Return(nil, errors.New("not found"))

		require.NoError(t, svc.ReloadGroupMappings(context.Background()))
		assert.Empty(t, svc.GetGroupMappings())
	})

	t.Run("ReloadGroupMappings fails on bad JSON", func(t *testing.T) {
		svc, _, _, settingRepo, ctrl := newTestOIDCService(t)
		defer ctrl.Finish()

		settingRepo.EXPECT().Get(gomock.Any(), "oidc_group_mappings").
			Return(&domain.Setting{Value: "not-json"}, nil)

		err := svc.ReloadGroupMappings(context.Background())
		require.Error(t, err)
	})

	t.Run("GetGroupMappings returns empty slice (not nil) when unset", func(t *testing.T) {
		svc, _, _, _, ctrl := newTestOIDCService(t)
		defer ctrl.Finish()
		got := svc.GetGroupMappings()
		assert.NotNil(t, got)
		assert.Empty(t, got)
	})

	t.Run("SetGroupMappings persists and updates cache", func(t *testing.T) {
		svc, _, _, settingRepo, ctrl := newTestOIDCService(t)
		defer ctrl.Finish()

		var captured string
		settingRepo.EXPECT().Set(gomock.Any(), "oidc_group_mappings", gomock.Any()).
			DoAndReturn(func(_ context.Context, _, value string) error {
				captured = value
				return nil
			})

		mappings := []domain.OIDCGroupMapping{{OIDCGroup: "g1", Role: "owner"}}
		require.NoError(t, svc.SetGroupMappings(context.Background(), mappings))

		var decoded []domain.OIDCGroupMapping
		require.NoError(t, json.Unmarshal([]byte(captured), &decoded))
		assert.Equal(t, mappings, decoded)
		assert.Equal(t, mappings, svc.GetGroupMappings())
	})

	t.Run("SetGroupMappings propagates store errors", func(t *testing.T) {
		svc, _, _, settingRepo, ctrl := newTestOIDCService(t)
		defer ctrl.Finish()
		settingRepo.EXPECT().Set(gomock.Any(), "oidc_group_mappings", gomock.Any()).
			Return(errors.New("boom"))

		err := svc.SetGroupMappings(context.Background(), nil)
		require.Error(t, err)
	})
}

func TestOIDCService_ConfigFlags(t *testing.T) {
	svc, _, _, _, ctrl := newTestOIDCService(t)
	defer ctrl.Finish()

	assert.True(t, svc.IsEnabled())
	assert.True(t, svc.AllowMagicCode())

	svc.config.Enabled = false
	svc.config.AllowMagicCode = false
	assert.False(t, svc.IsEnabled())
	assert.False(t, svc.AllowMagicCode())
}

func TestOIDCService_StateCacheLifecycle(t *testing.T) {
	// Stop must be safe to call even if cache is nil (defensive).
	svc := &OIDCService{}
	svc.Stop() // must not panic

	svc2, _, _, _, ctrl := newTestOIDCService(t)
	defer ctrl.Finish()

	// Insert a state and verify retrieval/expiry semantics through cache directly.
	st := &domain.OIDCState{State: "s1", Nonce: "n1", CreatedAt: time.Now()}
	svc2.stateCache.Set("s1", st, 1*time.Minute)
	got, ok := svc2.stateCache.Get("s1")
	require.True(t, ok)
	assert.Equal(t, st, got)

	svc2.stateCache.Delete("s1")
	_, ok = svc2.stateCache.Get("s1")
	assert.False(t, ok)
}
