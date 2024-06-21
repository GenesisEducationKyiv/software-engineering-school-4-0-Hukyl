//go:build integration

package tests_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Hukyl/genesis-kma-school-entry/database"
	"github.com/Hukyl/genesis-kma-school-entry/mail"
	mailCfg "github.com/Hukyl/genesis-kma-school-entry/mail/config"
	"github.com/Hukyl/genesis-kma-school-entry/models"
	"github.com/Hukyl/genesis-kma-school-entry/rate/fetchers"
	"github.com/Hukyl/genesis-kma-school-entry/server"
	serverCfg "github.com/Hukyl/genesis-kma-school-entry/server/config"
	"github.com/Hukyl/genesis-kma-school-entry/server/notifications"
	"github.com/Hukyl/genesis-kma-school-entry/server/notifications/message"
	"github.com/Hukyl/genesis-kma-school-entry/server/service"
	"github.com/Hukyl/genesis-kma-school-entry/settings"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRateServiceFetchRate_Success(t *testing.T) {
	// Arrange
	rateRepo := models.NewRateRepository(database.SetUpTest(t, &models.Rate{}))
	baseFetcher := fetchers.NewBaseFetcher()
	nbuFetcher := fetchers.NewNBURateFetcher()
	nbuFetcher.SetNext(baseFetcher)
	rateFetcher := service.NewRateService(rateRepo, nbuFetcher)
	// Act
	result, err := rateFetcher.FetchRate("USD", "UAH")
	// Assert
	assert.NoError(t, err)
	assert.Equal(t, "USD", result.CurrencyFrom)
	assert.Equal(t, "UAH", result.CurrencyTo)
	assert.NotZero(t, result.Rate)
}

func TestUserNotificationsRecipients(t *testing.T) {
	// Arrange
	ctx := context.Background()
	db := database.SetUpTest(t, &models.User{}, &models.Rate{})
	repo := models.NewUserRepository(db)

	rateService := service.NewRateService(
		models.NewRateRepository(db),
		fetchers.NewNBURateFetcher(),
	)
	smtpmockServer := mail.MockSMTPServer(t)
	emailClient := mail.Client{
		Config: mailCfg.Config{
			FromEmail:    "example@gmail.com",
			SMTPHost:     mail.Localhost,
			SMTPPort:     fmt.Sprint(smtpmockServer.PortNumber()),
			SMTPUser:     "user",
			SMTPPassword: "password",
		},
	}
	messageFormatter := message.PlainRateMessage{}
	notifier := notifications.NewUsersNotifier(
		&emailClient, rateService, repo, &messageFormatter,
	)
	users := []models.User{
		{Email: "test1@gmail.com"},
		{Email: "test2@gmail.com"},
	}
	for _, user := range users {
		if err := repo.Create(&user); err != nil {
			t.Fatalf("failed to create user: %v", err)
		}
	}
	// Act
	notifier.Notify(ctx)
	// Assert
	messages := smtpmockServer.Messages()
	assert.Len(t, messages, len(users))
	for i, user := range users {
		msg := messages[i]
		assert.Len(t, msg.RcpttoRequestResponse(), 1)
		rcpt := msg.RcpttoRequestResponse()[0][0]
		assert.Contains(t, rcpt, user.Email)
		assert.Contains(t, msg.MailfromRequest(), emailClient.Config.FromEmail)
	}
}

func TestSubscribeUser_Success(t *testing.T) {
	// Arrange
	db := database.SetUpTest(t, &models.User{})
	user := &models.User{Email: "example@gmail.com"}
	repo := models.NewUserRepository(db)
	engine := server.NewEngine(server.Client{
		Config:   serverCfg.Config{Port: "8080"},
		UserRepo: repo,
	})
	// Act
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, server.SubscribePath, nil)
	req.PostForm = map[string][]string{
		"email": {user.Email},
	}
	engine.ServeHTTP(rr, req)
	// Assert
	assert.Equal(t, http.StatusOK, rr.Code)
	exists, err := repo.Exists(user)
	assert.Nil(t, err)
	assert.True(t, exists)
}

func TestSubscribeUser_Conflict(t *testing.T) {
	user := &models.User{Email: "example@gmail.com"}
	// Arrange
	repo := models.NewUserRepository(database.SetUpTest(t, &models.User{}))
	err := repo.Create(user)
	assert.Nil(t, err)
	engine := server.NewEngine(server.Client{
		Config:   serverCfg.Config{Port: "8080"},
		UserRepo: repo,
	})
	// Act
	req := httptest.NewRequest(http.MethodPost, server.SubscribePath, nil)
	req.PostForm = map[string][]string{
		"email": {user.Email},
	}
	rr := httptest.NewRecorder()
	engine.ServeHTTP(rr, req)
	// Assert
	assert.Equal(t, http.StatusConflict, rr.Code)
	exists, err := repo.Exists(user)
	assert.Nil(t, err)
	assert.True(t, exists)
}

func TestMain(m *testing.M) {
	gin.SetMode(gin.ReleaseMode)
	_ = settings.InitSettings()
	m.Run()
}
