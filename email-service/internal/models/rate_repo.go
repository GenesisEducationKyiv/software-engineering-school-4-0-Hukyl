package models

import "gorm.io/gorm"

type DB interface {
	Connection() *gorm.DB
}

type RateRepository struct {
	db DB
}

func (r *RateRepository) Create(rate *Rate) error {
	return r.db.Connection().Create(rate).Error
}

func (r *RateRepository) Latest(ccFrom, ccTo string) (*Rate, error) {
	rate := &Rate{}
	err := r.db.Connection().
		Where("cc_from = ? AND cc_to = ?", ccFrom, ccTo).
		Order("created DESC").
		First(rate).
		Error
	return rate, err
}

func NewRateRepository(db DB) *RateRepository {
	return &RateRepository{db: db}
}
