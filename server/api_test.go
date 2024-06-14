package server_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/Hukyl/genesis-kma-school-entry/database"
	"github.com/Hukyl/genesis-kma-school-entry/models"
	"github.com/Hukyl/genesis-kma-school-entry/rate"
	"github.com/Hukyl/genesis-kma-school-entry/server"
	serverCfg "github.com/Hukyl/genesis-kma-school-entry/server/config"
	"github.com/Hukyl/genesis-kma-school-entry/settings"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestEmptyContext(t *testing.T) {
	engine := server.NewEngine(context.Background())
	req, err := http.NewRequest("GET", "/rate", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	engine.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestGetRate(t *testing.T) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, settings.DebugKey, true)
	apiClient := server.Client{
		Config:      serverCfg.NewFromEnv(),
		RateFetcher: rate.NewNBURateFetcher(),
	}
	ctx = context.WithValue(ctx, settings.APIClientKey, apiClient)
	engine := server.NewEngine(ctx)

	// FIXME: tests should not depend on external services
	req, err := http.NewRequest("GET", "/rate", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	engine.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
	rate, err := strconv.ParseFloat(rr.Body.String(), 32)
	assert.Nil(t, err)
	assert.Greater(t, rate, 0.0)
}

func TestSubscribeUserNoEmail(t *testing.T) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, settings.DebugKey, true)
	apiClient := server.Client{
		Config:      serverCfg.NewFromEnv(),
		RateFetcher: rate.NewNBURateFetcher(),
	}
	ctx = context.WithValue(ctx, settings.APIClientKey, apiClient)
	engine := server.NewEngine(ctx)

	req, err := http.NewRequest("POST", "/subscribe", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	engine.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestSubscribeUser(t *testing.T) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, settings.DebugKey, true)
	apiClient := server.Client{
		Config:      serverCfg.NewFromEnv(),
		RateFetcher: rate.NewNBURateFetcher(),
		UserRepo:    *models.NewUserRepository(database.SetUpTest(t, &models.User{})),
	}
	ctx = context.WithValue(ctx, settings.APIClientKey, apiClient)
	engine := server.NewEngine(ctx)

	req, err := http.NewRequest("POST", "/subscribe", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.PostForm = map[string][]string{
		"email": {"example@gmail.com"},
	}
	rr := httptest.NewRecorder()
	engine.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestSubscribeUserAlreadySubscribed(t *testing.T) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, settings.DebugKey, true)
	apiClient := server.Client{
		Config:      serverCfg.NewFromEnv(),
		RateFetcher: rate.NewNBURateFetcher(),
		UserRepo:    *models.NewUserRepository(database.SetUpTest(t, &models.User{})),
	}
	ctx = context.WithValue(ctx, settings.APIClientKey, apiClient)
	engine := server.NewEngine(ctx)

	req, err := http.NewRequest("POST", "/subscribe", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.PostForm = map[string][]string{
		"email": {"example@gmail.com"},
	}
	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, req)
	assert.Equal(t, http.StatusOK, recorder.Code)

	recorder = httptest.NewRecorder()
	engine.ServeHTTP(recorder, req)
	assert.Equal(t, http.StatusConflict, recorder.Code)
}

func TestMain(m *testing.M) {
	gin.SetMode(gin.ReleaseMode)
	_ = settings.InitSettings()
	m.Run()
}
