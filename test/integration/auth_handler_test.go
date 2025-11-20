package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"

	"github.com/example/user-service/internal/domain"
	"github.com/example/user-service/internal/ports/http/handlers"
	"github.com/example/user-service/internal/service"
)

type authServiceStub struct {
	lastOAuthInfo *service.OAuthUserInfo
}

func (authServiceStub) StartSignup(ctx context.Context, traceID, email, password string) (string, error) {
	return "uuid-1", nil
}

func (authServiceStub) VerifySignup(ctx context.Context, traceID, uuid, code string) (*domain.User, *service.Tokens, error) {
	return &domain.User{ID: "user-1", Email: "user@example.com"}, &service.Tokens{AccessToken: "token"}, nil
}

func (authServiceStub) SignIn(ctx context.Context, traceID, email, password string) (*domain.User, *service.Tokens, error) {
	return &domain.User{ID: "user-1", Email: "user@example.com"}, &service.Tokens{AccessToken: "token"}, nil
}

func (s *authServiceStub) HandleOAuthCallback(ctx context.Context, traceID string, info service.OAuthUserInfo) (*domain.User, *service.Tokens, error) {
	s.lastOAuthInfo = &info
	return &domain.User{ID: "user-1", Email: "user@example.com"}, &service.Tokens{AccessToken: "token"}, nil
}

func TestAuthHandlerSignup(t *testing.T) {
	e := echo.New()
	handler := handlers.NewAuthHandler(&authServiceStub{})

	reqBody, _ := json.Marshal(map[string]string{"email": "user@example.com", "password": "password123"})
	req := httptest.NewRequest(http.MethodPost, "/auth/signup", bytes.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.Signup(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusAccepted, rec.Code)
}

func TestAuthHandlerOAuthCallback(t *testing.T) {
	e := echo.New()
	stub := &authServiceStub{}
	handler := handlers.NewAuthHandler(stub)

	body := map[string]interface{}{
		"provider_type":    "google",
		"provider_user_id": "abc123",
		"email":            "user@example.com",
		"display_name":     "OAuth User",
		"metadata":         map[string]interface{}{"locale": "en"},
	}

	payload, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/auth/oauth/callback", bytes.NewReader(payload))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.HandleOAuthCallback(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.NotNil(t, stub.lastOAuthInfo)
	assert.Equal(t, "google", stub.lastOAuthInfo.ProviderType)
	assert.Equal(t, "abc123", stub.lastOAuthInfo.ProviderUserID)
	assert.Equal(t, "user@example.com", stub.lastOAuthInfo.Email)
}
