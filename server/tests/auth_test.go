package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/xrlnewman/homeflow-admin/server/internal/config"
	"github.com/xrlnewman/homeflow-admin/server/internal/platform/store"
	"github.com/xrlnewman/homeflow-admin/server/internal/transport/httpapi"
)

func TestHealthEndpointReportsDependencies(t *testing.T) {
	r := httpapi.NewRouter(config.Config{JWTSecret: "test-secret"}, store.NewMemoryStore())
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	res := httptest.NewRecorder()
	r.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.Code)
	}
}

func TestProtectedEndpointRequiresBearerToken(t *testing.T) {
	r := httpapi.NewRouter(config.Config{JWTSecret: "test-secret"}, store.NewMemoryStore())
	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	res := httptest.NewRecorder()
	r.ServeHTTP(res, req)
	if res.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", res.Code)
	}
}

func TestLoginReturnsBearerTokenAndMe(t *testing.T) {
	r := httpapi.NewRouter(config.Config{JWTSecret: "test-secret"}, store.NewMemoryStore())
	body, _ := json.Marshal(map[string]string{"phone": "13800000000", "password": "demo123456"})
	loginReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
	loginReq.Header.Set("Content-Type", "application/json")
	loginRes := httptest.NewRecorder()
	r.ServeHTTP(loginRes, loginReq)
	if loginRes.Code != http.StatusOK {
		t.Fatalf("expected login 200, got %d: %s", loginRes.Code, loginRes.Body.String())
	}
	var envelope struct {
		Data struct {
			AccessToken string `json:"accessToken"`
		} `json:"data"`
	}
	if err := json.Unmarshal(loginRes.Body.Bytes(), &envelope); err != nil || envelope.Data.AccessToken == "" {
		t.Fatalf("expected access token, body=%s", loginRes.Body.String())
	}
	meReq := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	meReq.Header.Set("Authorization", "Bearer "+envelope.Data.AccessToken)
	meRes := httptest.NewRecorder()
	r.ServeHTTP(meRes, meReq)
	if meRes.Code != http.StatusOK {
		t.Fatalf("expected me 200, got %d", meRes.Code)
	}
}

func TestCustomerCannotAssignOrder(t *testing.T) {
	r := httpapi.NewRouter(config.Config{JWTSecret: "test-secret"}, store.NewMemoryStore())
	body, _ := json.Marshal(map[string]string{"phone": "13800000000", "password": "demo123456"})
	loginReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
	loginReq.Header.Set("Content-Type", "application/json")
	loginRes := httptest.NewRecorder()
	r.ServeHTTP(loginRes, loginReq)
	var envelope struct {
		Data struct {
			AccessToken string `json:"accessToken"`
		} `json:"data"`
	}
	_ = json.Unmarshal(loginRes.Body.Bytes(), &envelope)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/orders/order-1/assign", bytes.NewBufferString(`{"technicianId":"tech-demo"}`))
	req.Header.Set("Authorization", "Bearer "+envelope.Data.AccessToken)
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	r.ServeHTTP(res, req)
	if res.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", res.Code)
	}
}
