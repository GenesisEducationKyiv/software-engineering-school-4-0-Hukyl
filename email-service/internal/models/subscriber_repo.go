package models

type SubscriberRepository struct {
	db DB
}

func NewSubscriberRepository(db DB) *SubscriberRepository {
	return &SubscriberRepository{db: db}
}

func (r *SubscriberRepository) Create(subscriber *Subscriber) error {
	return r.db.Connection().Create(subscriber).Error
}

func (r *SubscriberRepository) FindAll() ([]Subscriber, error) {
	var subscribers []Subscriber
	err := r.db.Connection().Find(&subscribers).Error
	return subscribers, err
}

func (r *SubscriberRepository) Exists(subscriber *Subscriber) (bool, error) {
	var count int64
	err := r.db.Connection().Model(&Subscriber{}).Where("email = ?", subscriber.Email).Count(
		&count,
	).Error
	return count > 0, err
}

func (r *SubscriberRepository) Delete(subscriber *Subscriber) error {
	conn := r.db.Connection()
	return conn.Where("email = ?", subscriber.Email).Delete(subscriber).Error
}
