package server_test

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/Hukyl/genesis-kma-school-entry/server"
	"github.com/Hukyl/genesis-kma-school-entry/utils"
)

func TestGetRate(t *testing.T) {
	// FIXME: tests should not depend on external services
	r := server.ApiEngine()
	req, err := http.NewRequest("GET", "/rate", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
	rate, err := strconv.ParseFloat(rr.Body.String(), 32)
	assert.Nil(t, err)
	assert.Greater(t, rate, 0.0)
}

func TestSubscribeUserNoEmail(t *testing.T) {
	r := server.ApiEngine()
	req, err := http.NewRequest("POST", "/subscribe", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestSubscribeUser(t *testing.T) {
	utils.SetUpTestDB()
	defer utils.TearDownTestDB()

	r := server.ApiEngine()
	req, err := http.NewRequest("POST", "/subscribe", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.PostForm = map[string][]string{
		"email": {"example@gmail.com"},
	}
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestSubscribeUserAlreadySubscribed(t *testing.T) {
	utils.SetUpTestDB()
	defer utils.TearDownTestDB()

	r := server.ApiEngine()

	req, err := http.NewRequest("POST", "/subscribe", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.PostForm = map[string][]string{
		"email": {"example@gmail.com"},
	}
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)

	rr = httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusConflict, rr.Code)
}

func TestMain(m *testing.M) {
	utils.SetUpTestDB()
	defer utils.TearDownTestDB()
	gin.SetMode(gin.ReleaseMode)
	m.Run()
}
