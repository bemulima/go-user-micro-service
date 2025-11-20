package unit

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/example/user-service/config"
	"github.com/example/user-service/internal/domain"
	"github.com/example/user-service/internal/ports/tarantool"
	"github.com/example/user-service/internal/service"
	pkglog "github.com/example/user-service/pkg/log"
)

type fakeUserRepo struct {
	users map[string]*domain.User
}

func newFakeUserRepo() *fakeUserRepo {
	return &fakeUserRepo{users: map[string]*domain.User{}}
}

func (f *fakeUserRepo) Create(ctx context.Context, user *domain.User) error {
	user.ID = "user-1"
	f.users[strings.ToLower(user.Email)] = user
	return nil
}

func (f *fakeUserRepo) Update(ctx context.Context, user *domain.User) error {
	f.users[strings.ToLower(user.Email)] = user
	return nil
}

func (f *fakeUserRepo) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	if user, ok := f.users[strings.ToLower(email)]; ok {
		return user, nil
	}
	return nil, gorm.ErrRecordNotFound
}

func (f *fakeUserRepo) FindByID(ctx context.Context, id string) (*domain.User, error) {
	for _, user := range f.users {
		if user.ID == id {
			return user, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (f *fakeUserRepo) Delete(ctx context.Context, id string) error { return nil }
func (f *fakeUserRepo) List(ctx context.Context, offset, limit int) ([]domain.User, int64, error) {
	return nil, 0, nil
}

type fakeProfileRepo struct {
	profiles map[string]*domain.UserProfile
}

func newFakeProfileRepo() *fakeProfileRepo {
	return &fakeProfileRepo{profiles: map[string]*domain.UserProfile{}}
}

func (f *fakeProfileRepo) Create(ctx context.Context, profile *domain.UserProfile) error {
	profile.ID = "profile-1"
	f.profiles[profile.UserID] = profile
	return nil
}

func (f *fakeProfileRepo) Update(ctx context.Context, profile *domain.UserProfile) error {
	f.profiles[profile.UserID] = profile
	return nil
}

func (f *fakeProfileRepo) FindByUserID(ctx context.Context, userID string) (*domain.UserProfile, error) {
	if profile, ok := f.profiles[userID]; ok {
		return profile, nil
	}
	return nil, errors.New("not found")
}

type fakeProviderRepo struct {
	providers map[string]*domain.UserProvider
}

func newFakeProviderRepo() *fakeProviderRepo {
	return &fakeProviderRepo{providers: map[string]*domain.UserProvider{}}
}

func (f *fakeProviderRepo) key(providerType, providerUserID string) string {
	return providerType + ":" + providerUserID
}

func (f *fakeProviderRepo) Create(ctx context.Context, provider *domain.UserProvider) error {
	provider.ID = "provider-" + provider.ProviderUserID
	f.providers[f.key(provider.ProviderType, provider.ProviderUserID)] = provider
	return nil
}

func (f *fakeProviderRepo) Update(ctx context.Context, provider *domain.UserProvider) error {
	f.providers[f.key(provider.ProviderType, provider.ProviderUserID)] = provider
	return nil
}

func (f *fakeProviderRepo) Delete(ctx context.Context, id string) error { return nil }

func (f *fakeProviderRepo) FindByProvider(ctx context.Context, providerType, providerUserID string) (*domain.UserProvider, error) {
	if provider, ok := f.providers[f.key(providerType, providerUserID)]; ok {
		return provider, nil
	}
	return nil, gorm.ErrRecordNotFound
}

func (f *fakeProviderRepo) FindByUserID(ctx context.Context, userID string) ([]domain.UserProvider, error) {
	var result []domain.UserProvider
	for _, provider := range f.providers {
		if provider.UserID == userID {
			result = append(result, *provider)
		}
	}
	return result, nil
}

type fakeTarantool struct {
	email    string
	password string
}

func (f *fakeTarantool) StartRegistration(ctx context.Context, email, password string) (string, error) {
	f.email = email
	f.password = password
	return "uuid-1", nil
}

func (f *fakeTarantool) VerifyRegistration(ctx context.Context, uuid, code string) (*tarantool.VerificationResult, error) {
	return &tarantool.VerificationResult{Email: f.email, Password: f.password}, nil
}

func (f *fakeTarantool) StartEmailChange(ctx context.Context, userID, email string) (string, error) {
	return "uuid-change", nil
}

func (f *fakeTarantool) VerifyEmailChange(ctx context.Context, uuid, code string) (*tarantool.VerificationResult, error) {
	return &tarantool.VerificationResult{Email: "new@example.com"}, nil
}

type fakePublisher struct{}

func (fakePublisher) Publish(ctx context.Context, routingKey string, payload interface{}) error {
	return nil
}
func (fakePublisher) Close() error { return nil }

func TestAuthService_StartSignup(t *testing.T) {
	cfg := &config.Config{JWTSecret: "secret", JWTTTLMinutes: time.Minute, JWTRefreshTTLMinutes: time.Hour}
	signer, err := service.NewJWTSigner(cfg)
	require.NoError(t, err)
	users := newFakeUserRepo()
	profiles := newFakeProfileRepo()
	providers := newFakeProviderRepo()
	tarantoolClient := &fakeTarantool{}
	auth := service.NewAuthService(cfg, pkglog.New("test"), users, profiles, providers, tarantoolClient, fakePublisher{}, signer)

	uuid, err := auth.StartSignup(context.Background(), "trace-1", "user@example.com", "password123")
	require.NoError(t, err)
	assert.Equal(t, "uuid-1", uuid)
}

func TestAuthService_VerifySignup(t *testing.T) {
	cfg := &config.Config{JWTSecret: "secret", JWTTTLMinutes: time.Minute, JWTRefreshTTLMinutes: time.Hour}
	signer, err := service.NewJWTSigner(cfg)
	require.NoError(t, err)
	users := newFakeUserRepo()
	profiles := newFakeProfileRepo()
	providers := newFakeProviderRepo()
	tarantoolClient := &fakeTarantool{email: "user@example.com", password: "password123"}
	auth := service.NewAuthService(cfg, pkglog.New("test"), users, profiles, providers, tarantoolClient, fakePublisher{}, signer)

	user, tokens, err := auth.VerifySignup(context.Background(), "trace-1", "uuid-1", "code")
	require.NoError(t, err)
	assert.NotNil(t, user)
	assert.NotNil(t, tokens)
	assert.NotEmpty(t, tokens.AccessToken)
	assert.NotNil(t, user.PasswordHash)
}

func TestAuthService_HandleOAuthCallback_CreateAndLink(t *testing.T) {
	cfg := &config.Config{JWTSecret: "secret", JWTTTLMinutes: time.Minute, JWTRefreshTTLMinutes: time.Hour}
	signer, err := service.NewJWTSigner(cfg)
	require.NoError(t, err)
	users := newFakeUserRepo()
	profiles := newFakeProfileRepo()
	providers := newFakeProviderRepo()
	tarantoolClient := &fakeTarantool{}
	auth := service.NewAuthService(cfg, pkglog.New("test"), users, profiles, providers, tarantoolClient, fakePublisher{}, signer)

	displayName := "OAuth User"
	user, tokens, err := auth.HandleOAuthCallback(context.Background(), "trace-1", service.OAuthUserInfo{
		ProviderType:   "google",
		ProviderUserID: "oauth-1",
		Email:          "oauth@example.com",
		DisplayName:    &displayName,
		Metadata:       map[string]interface{}{"locale": "en"},
	})

	require.NoError(t, err)
	assert.NotNil(t, user)
	assert.NotNil(t, tokens)
	assert.NotEmpty(t, tokens.AccessToken)

	provider, err := providers.FindByProvider(context.Background(), "google", "oauth-1")
	require.NoError(t, err)
	assert.Equal(t, user.ID, provider.UserID)
}
