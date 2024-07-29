package models

import (
	"time"

	"github.com/VictoriaMetrics/metrics"
)

var (
	subInsertTimeMetric = metrics.GetOrCreateHistogram(
		`database_subscriber_query_duration_seconds{type="insert"}`,
	)
	subDeleteTimeMetric = metrics.GetOrCreateHistogram(
		`database_subscriber_query_duration_seconds{type="delete"}`,
	)
	subSelectTimeMetric = metrics.GetOrCreateHistogram(
		`database_subscriber_query_duration_seconds{type="select"}`,
	)
)

type SubscriberRepository struct {
	db DB
}

func NewSubscriberRepository(db DB) *SubscriberRepository {
	return &SubscriberRepository{db: db}
}

func (r *SubscriberRepository) Create(subscriber *Subscriber) error {
	startTime := time.Now()
	err := r.db.Connection().Create(subscriber).Error
	subInsertTimeMetric.UpdateDuration(startTime)
	return err
}

func (r *SubscriberRepository) FindAll() ([]Subscriber, error) {
	startTime := time.Now()
	var subscribers []Subscriber
	err := r.db.Connection().Find(&subscribers).Error
	subSelectTimeMetric.UpdateDuration(startTime)
	return subscribers, err
}

func (r *SubscriberRepository) Exists(subscriber *Subscriber) (bool, error) {
	startTime := time.Now()
	var count int64
	err := r.db.Connection().Model(&Subscriber{}).Where("email = ?", subscriber.Email).Count(
		&count,
	).Error
	subSelectTimeMetric.UpdateDuration(startTime)
	return count > 0, err
}

func (r *SubscriberRepository) Delete(subscriber *Subscriber) error {
	startTime := time.Now()
	conn := r.db.Connection()
	err := conn.Where("email = ?", subscriber.Email).Delete(subscriber).Error
	subDeleteTimeMetric.UpdateDuration(startTime)
	return err
}
