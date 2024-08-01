package models

import (
	"time"

	"github.com/VictoriaMetrics/metrics"
	"gorm.io/gorm"
)

type DB interface {
	Connection() *gorm.DB
}

var (
	insertTimeMetric = metrics.GetOrCreateHistogram(
		`database_user_query_duration_seconds{type="insert"}`,
	)
	deleteTimeMetric = metrics.GetOrCreateHistogram(
		`database_user_query_duration_seconds{type="delete"}`,
	)
	selectTimeMetric = metrics.GetOrCreateHistogram(
		`database_user_query_duration_seconds{type="select"}`,
	)
)

type UserRepository struct {
	db DB
}

func NewUserRepository(db DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(user *User) error {
	startTime := time.Now()
	err := r.db.Connection().Create(user).Error
	insertTimeMetric.UpdateDuration(startTime)
	return err
}

func (r *UserRepository) FindAll() ([]User, error) {
	startTime := time.Now()
	var users []User
	err := r.db.Connection().Find(&users).Error
	selectTimeMetric.UpdateDuration(startTime)
	return users, err
}

func (r *UserRepository) Exists(user *User) (bool, error) {
	startTime := time.Now()
	var count int64
	err := r.db.Connection().Model(&User{}).Where("email = ?", user.Email).Count(
		&count,
	).Error
	selectTimeMetric.UpdateDuration(startTime)
	return count > 0, err
}

func (r *UserRepository) Delete(user *User) error {
	startTime := time.Now()
	conn := r.db.Connection()
	err := conn.Where("email = ?", user.Email).Delete(user).Error
	deleteTimeMetric.UpdateDuration(startTime)
	return err
}
