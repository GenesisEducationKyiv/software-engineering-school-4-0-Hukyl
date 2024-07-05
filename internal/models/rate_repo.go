package models

type RateRepository struct {
	db DB
}

func (r *RateRepository) Create(rate *Rate) error {
	return r.db.Connection().Create(rate).Error
}

func NewRateRepository(db DB) *RateRepository {
	return &RateRepository{db: db}
}
