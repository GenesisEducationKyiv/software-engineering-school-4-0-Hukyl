package server_test

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/Hukyl/genesis-kma-school-entry/models"
	"github.com/Hukyl/genesis-kma-school-entry/rate"
	"github.com/Hukyl/genesis-kma-school-entry/server"
	serverCfg "github.com/Hukyl/genesis-kma-school-entry/server/config"
	"github.com/Hukyl/genesis-kma-school-entry/settings"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockRateFetcher struct {
	mock.Mock
}

func (m *mockRateFetcher) FetchRate(from, to string) (rate.Rate, error) {
	args := m.Called(from, to)
	return args.Get(0).(rate.Rate), args.Error(1)
}

type mockUserRepository struct {
	mock.Mock
}

func (m *mockUserRepository) FindAll() ([]models.User, error) {
	args := m.Called()
	return args.Get(0).([]models.User), args.Error(1)
}

func (m *mockUserRepository) Create(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *mockUserRepository) Exists(user *models.User) (bool, error) {
	args := m.Called(user)
	return args.Bool(0), args.Error(1)
}

func TestGetRate(t *testing.T) {
	mockFetcher := new(mockRateFetcher)
	mockedRate := rate.Rate{Rate: 27.5}
	mockFetcher.On("FetchRate", "USD", "UAH").Return(mockedRate, nil)
	engine := server.NewEngine(server.Client{
		Config:      serverCfg.Config{Port: "8080"},
		RateFetcher: mockFetcher,
	})

	req := httptest.NewRequest(http.MethodGet, server.RatePath, nil)
	rr := httptest.NewRecorder()
	engine.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
	rate, err := strconv.ParseFloat(rr.Body.String(), 32)
	assert.Nil(t, err)
	assert.EqualValues(t, rate, mockedRate.Rate)
}

func TestSubscribeUserNoEmail(t *testing.T) {
	engine := server.NewEngine(server.Client{
		Config:      serverCfg.Config{Port: "8080"},
		RateFetcher: &mockRateFetcher{},
	})

	req := httptest.NewRequest(http.MethodPost, server.SubscribePath, nil)
	rr := httptest.NewRecorder()
	engine.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestSubscribeUser(t *testing.T) {
	mockRepo := new(mockUserRepository)
	mockRepo.On("Exists", mock.Anything).Return(false, nil).Once()
	mockRepo.On("Create", mock.Anything).Return(nil).Once()
	engine := server.NewEngine(server.Client{
		Config:   serverCfg.Config{Port: "8080"},
		UserRepo: mockRepo,
	})

	req := httptest.NewRequest(http.MethodPost, server.SubscribePath, nil)
	req.PostForm = map[string][]string{
		"email": {"example@gmail.com"},
	}
	rr := httptest.NewRecorder()
	engine.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestSubscribeUserAlreadySubscribed(t *testing.T) {
	mockRepo := new(mockUserRepository)
	mockRepo.On("Exists", mock.Anything).Return(true, nil).Once()
	engine := server.NewEngine(server.Client{
		Config:   serverCfg.Config{Port: "8080"},
		UserRepo: mockRepo,
	})

	req := httptest.NewRequest(http.MethodPost, server.SubscribePath, nil)
	req.PostForm = map[string][]string{
		"email": {"example@gmail.com"},
	}
	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, req)
	assert.Equal(t, http.StatusConflict, recorder.Code)
}

func TestMain(m *testing.M) {
	gin.SetMode(gin.ReleaseMode)
	_ = settings.InitSettings()
	m.Run()
}
