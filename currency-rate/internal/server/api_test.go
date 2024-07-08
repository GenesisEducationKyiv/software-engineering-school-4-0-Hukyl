package server_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/currency-rate/internal/models"
	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/currency-rate/internal/rate"
	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/currency-rate/internal/server"
	serverCfg "github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/currency-rate/internal/server/config"
	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/pkg/settings"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type (
	mockUserRepository struct {
		mock.Mock
	}

	mockRateService struct {
		mock.Mock
	}
)

func (m *mockRateService) FetchRate(ctx context.Context, from, to string) (rate.Rate, error) {
	args := m.Called(ctx, from, to)
	return args.Get(0).(rate.Rate), args.Error(1)
}

func (m *mockUserRepository) FindAll() ([]models.User, error) {
	args := m.Called()
	return args.Get(0).([]models.User), args.Error(1)
}

func (m *mockUserRepository) Create(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *mockUserRepository) Delete(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *mockUserRepository) Exists(user *models.User) (bool, error) {
	args := m.Called(user)
	return args.Bool(0), args.Error(1)
}

func TestGetRate(t *testing.T) {
	mockService := new(mockRateService)
	mockedRate := &rate.Rate{Rate: 27.5}
	mockService.On("FetchRate", mock.Anything, "USD", "UAH").Return(mockedRate, nil)
	engine := server.NewEngine(server.Client{
		Config:      serverCfg.Config{Port: "8080"},
		RateService: mockService,
	})

	req := httptest.NewRequest(http.MethodGet, server.RatePath, nil)
	rr := httptest.NewRecorder()
	engine.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
	rate, err := strconv.ParseFloat(rr.Body.String(), 32)
	require.NoError(t, err)
	assert.InDelta(t, rate, mockedRate.Rate, 0.001)
}

func TestSubscribeUserNoEmail(t *testing.T) {
	engine := server.NewEngine(server.Client{
		Config:      serverCfg.Config{Port: "8080"},
		RateService: &mockRateService{},
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

func TestUnsubscribeUserNoEmail(t *testing.T) {
	engine := server.NewEngine(server.Client{
		Config:      serverCfg.Config{Port: "8080"},
		RateService: &mockRateService{},
	})

	req := httptest.NewRequest(http.MethodPost, server.UnsubscribePath, nil)
	rr := httptest.NewRecorder()
	engine.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestUnsubscribeUser(t *testing.T) {
	mockRepo := new(mockUserRepository)
	mockRepo.On("Exists", mock.Anything).Return(true, nil).Once()
	mockRepo.On("Delete", mock.Anything).Return(nil).Once()
	engine := server.NewEngine(server.Client{
		Config:   serverCfg.Config{Port: "8080"},
		UserRepo: mockRepo,
	})

	req := httptest.NewRequest(http.MethodPost, server.UnsubscribePath, nil)
	req.PostForm = map[string][]string{
		"email": {"example@gmail.com"},
	}
	rr := httptest.NewRecorder()
	engine.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestUnsubscribeUserNotSubscribed(t *testing.T) {
	mockRepo := new(mockUserRepository)
	mockRepo.On("Exists", mock.Anything).Return(false, nil).Once()
	engine := server.NewEngine(server.Client{
		Config:   serverCfg.Config{Port: "8080"},
		UserRepo: mockRepo,
	})

	req := httptest.NewRequest(http.MethodPost, server.UnsubscribePath, nil)
	req.PostForm = map[string][]string{
		"email": {"example@gmail.com"},
	}
	rr := httptest.NewRecorder()
	engine.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusGone, rr.Code)
}

func TestMain(m *testing.M) {
	gin.SetMode(gin.ReleaseMode)
	_ = settings.InitSettings()
	m.Run()
}
