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
	rateInsertTimeMetric = metrics.GetOrCreateHistogram(
		`database_rate_query_duration_seconds{type="insert"}`,
	)
	rateSelectTimeMetric = metrics.GetOrCreateHistogram(
		`database_rate_query_duration_seconds{type="select"}`,
	)
)

type RateRepository struct {
	db DB
}

func (r *RateRepository) Create(rate *Rate) error {
	startTime := time.Now()
	err := r.db.Connection().Create(rate).Error
	rateInsertTimeMetric.UpdateDuration(startTime)
	return err
}

func (r *RateRepository) Latest(ccFrom, ccTo string) (*Rate, error) {
	startTime := time.Now()
	rate := &Rate{}
	err := r.db.Connection().
		Where("cc_from = ? AND cc_to = ?", ccFrom, ccTo).
		Order("created DESC").
		First(rate).
		Error
	rateSelectTimeMetric.UpdateDuration(startTime)
	return rate, err
}

func NewRateRepository(db DB) *RateRepository {
	return &RateRepository{db: db}
}
