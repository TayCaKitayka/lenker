package httpapi_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/lenker/lenker/services/panel-api/internal/admins"
	"github.com/lenker/lenker/services/panel-api/internal/auth"
	httpapi "github.com/lenker/lenker/services/panel-api/internal/http"
	"github.com/lenker/lenker/services/panel-api/internal/storage"
	"github.com/lenker/lenker/services/panel-api/internal/users"
	"golang.org/x/crypto/bcrypt"
)

func TestContractHealthResponseShape(t *testing.T) {
	router := newContractRouter(t, contractDeps{})

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)

	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", response.Code, response.Body.String())
	}
	body := decodeBody(t, response)
	data := body.object("data")
	if got := data.string("status"); got != "ok" {
		t.Fatalf("expected health status ok, got %q", got)
	}
	body.mustNotHave("error")
}

func TestContractAdminLoginSuccessShape(t *testing.T) {
	router := newContractRouter(t, contractDeps{})

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/admin/login", strings.NewReader(`{
		"email": "owner@example.com",
		"password": "secret"
	}`))

	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", response.Code, response.Body.String())
	}
	body := decodeBody(t, response)
	data := body.object("data")
	if got := data.object("admin").string("email"); got != "owner@example.com" {
		t.Fatalf("expected admin email, got %q", got)
	}
	if got := data.object("session").string("token"); got == "" {
		t.Fatalf("expected session token")
	}
	body.mustNotHave("error")
}

func TestContractAdminLoginInvalidCredentialsShape(t *testing.T) {
	router := newContractRouter(t, contractDeps{})

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/admin/login", strings.NewReader(`{
		"email": "owner@example.com",
		"password": "wrong"
	}`))

	router.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d: %s", response.Code, response.Body.String())
	}
	assertErrorEnvelope(t, response, "invalid_credentials")
}

func TestContractProtectedEndpointWithoutTokenShape(t *testing.T) {
	router := newContractRouter(t, contractDeps{})

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)

	router.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d: %s", response.Code, response.Body.String())
	}
	assertErrorEnvelope(t, response, "unauthorized")
}

func TestContractNotFoundShape(t *testing.T) {
	router := newContractRouter(t, contractDeps{
		users: &contractUsersRepository{findErr: storage.ErrNotFound},
	})

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/users/user-404", nil)
	request.Header.Set("Authorization", "Bearer test-token")

	router.ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d: %s", response.Code, response.Body.String())
	}
	assertErrorEnvelope(t, response, "not_found")
}

func TestContractValidationErrorShape(t *testing.T) {
	router := newContractRouter(t, contractDeps{})

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/users", strings.NewReader(`{
		"email": ""
	}`))
	request.Header.Set("Authorization", "Bearer test-token")

	router.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d: %s", response.Code, response.Body.String())
	}
	assertErrorEnvelope(t, response, "bad_request")
}

type contractDeps struct {
	users storage.UsersRepository
}

func newContractRouter(t *testing.T, deps contractDeps) http.Handler {
	t.Helper()
	adminsRepo := newContractAdminsRepository(t)
	if deps.users == nil {
		deps.users = &contractUsersRepository{}
	}

	adminSession := auth.NewSessionMiddleware(nil, adminsRepo)
	return httpapi.NewRouter(httpapi.RouterDeps{
		Auth: auth.NewHandler(nil, auth.NewService(adminsRepo, auth.NewPasswordVerifier())),
		Users: users.NewHandler(
			nil,
			deps.users,
			adminSession.RequireAdmin,
		),
	})
}

func newContractAdminsRepository(t *testing.T) *contractAdminsRepository {
	t.Helper()
	passwordHash, err := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("failed to create bcrypt hash: %v", err)
	}
	return &contractAdminsRepository{
		admin: admins.Admin{
			ID:               "admin-1",
			Email:            "owner@example.com",
			PasswordHash:     string(passwordHash),
			Status:           "active",
			TwoFactorEnabled: false,
			CreatedAt:        time.Now().UTC(),
			UpdatedAt:        time.Now().UTC(),
		},
	}
}

type contractAdminsRepository struct {
	admin admins.Admin
}

func (r *contractAdminsRepository) FindByEmail(ctx context.Context, email string) (admins.Admin, error) {
	if email != r.admin.Email {
		return admins.Admin{}, admins.ErrNotFound
	}
	return r.admin, nil
}

func (r *contractAdminsRepository) FindByActiveSessionTokenHash(ctx context.Context, tokenHash string, now time.Time) (admins.Admin, error) {
	if tokenHash == "" {
		return admins.Admin{}, admins.ErrNotFound
	}
	return r.admin, nil
}

func (r *contractAdminsRepository) CreateSession(ctx context.Context, adminID string, tokenHash string, expiresAt time.Time) (admins.Session, error) {
	return admins.Session{
		ID:        "session-1",
		AdminID:   adminID,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now().UTC(),
	}, nil
}

type contractUsersRepository struct {
	findErr error
}

func (r *contractUsersRepository) List(ctx context.Context) ([]storage.User, error) {
	return []storage.User{}, nil
}

func (r *contractUsersRepository) Create(ctx context.Context, input storage.CreateUserInput) (storage.User, error) {
	return storage.User{ID: "user-1", Email: input.Email, Status: "active", DisplayName: input.DisplayName}, nil
}

func (r *contractUsersRepository) FindByID(ctx context.Context, id string) (storage.User, error) {
	if r.findErr != nil {
		return storage.User{}, r.findErr
	}
	return storage.User{ID: id, Email: "user@example.com", Status: "active"}, nil
}

func (r *contractUsersRepository) Update(ctx context.Context, id string, input storage.UpdateUserInput) (storage.User, error) {
	return storage.User{ID: id, Email: "user@example.com", Status: "active"}, nil
}

func (r *contractUsersRepository) SetStatus(ctx context.Context, id string, status string) (storage.User, error) {
	return storage.User{ID: id, Email: "user@example.com", Status: status}, nil
}

type responseBody map[string]any

func decodeBody(t *testing.T, response *httptest.ResponseRecorder) responseBody {
	t.Helper()
	var body responseBody
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("response is not JSON: %v\n%s", err, response.Body.String())
	}
	return body
}

func (b responseBody) object(key string) responseBody {
	value, ok := b[key].(map[string]any)
	if !ok {
		return nil
	}
	return responseBody(value)
}

func (b responseBody) string(key string) string {
	value, _ := b[key].(string)
	return value
}

func (b responseBody) mustNotHave(key string) {
	if _, ok := b[key]; ok {
		panic("unexpected key " + key)
	}
}

func assertErrorEnvelope(t *testing.T, response *httptest.ResponseRecorder, code string) {
	t.Helper()
	body := decodeBody(t, response)
	errObj := body.object("error")
	if errObj == nil {
		t.Fatalf("expected error envelope: %s", response.Body.String())
	}
	if got := errObj.string("code"); got != code {
		t.Fatalf("expected error code %q, got %q: %s", code, got, response.Body.String())
	}
	if got := errObj.string("message"); got == "" {
		t.Fatalf("expected error message: %s", response.Body.String())
	}
	if _, ok := body["data"]; ok {
		t.Fatalf("did not expect data in error envelope: %s", response.Body.String())
	}
}
