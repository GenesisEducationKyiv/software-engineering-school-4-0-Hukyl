package server_test

import (
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

func TestGetRate(t *testing.T) {
	engine := server.NewEngine(server.Client{
		Config:      serverCfg.NewFromEnv(),
		RateFetcher: rate.NewNBURateFetcher(),
	})

	// FIXME: tests should not depend on external services
	req := httptest.NewRequest(http.MethodGet, "/rate", nil)
	rr := httptest.NewRecorder()
	engine.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
	rate, err := strconv.ParseFloat(rr.Body.String(), 32)
	assert.Nil(t, err)
	assert.Greater(t, rate, 0.0)
}

func TestSubscribeUserNoEmail(t *testing.T) {
	engine := server.NewEngine(server.Client{
		Config:      serverCfg.NewFromEnv(),
		RateFetcher: rate.NewNBURateFetcher(),
	})

	req := httptest.NewRequest(http.MethodPost, "/subscribe", nil)
	rr := httptest.NewRecorder()
	engine.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestSubscribeUser(t *testing.T) {
	engine := server.NewEngine(server.Client{
		Config:      serverCfg.NewFromEnv(),
		RateFetcher: rate.NewNBURateFetcher(),
		UserRepo:    *models.NewUserRepository(database.SetUpTest(t, &models.User{})),
	})

	req := httptest.NewRequest(http.MethodPost, "/subscribe", nil)
	req.PostForm = map[string][]string{
		"email": {"example@gmail.com"},
	}
	rr := httptest.NewRecorder()
	engine.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestSubscribeUserAlreadySubscribed(t *testing.T) {
	engine := server.NewEngine(server.Client{
		Config:      serverCfg.NewFromEnv(),
		RateFetcher: rate.NewNBURateFetcher(),
		UserRepo:    *models.NewUserRepository(database.SetUpTest(t, &models.User{})),
	})

	req := httptest.NewRequest(http.MethodPost, "/subscribe", nil)
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
