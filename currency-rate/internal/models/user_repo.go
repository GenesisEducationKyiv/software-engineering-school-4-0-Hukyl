package models

import (
	"gorm.io/gorm"
)

type DB interface {
	Connection() *gorm.DB
}

type UserRepository struct {
	db DB
}

func NewUserRepository(db DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(user *User) error {
	return r.db.Connection().Create(user).Error
}

func (r *UserRepository) FindAll() ([]User, error) {
	var users []User
	err := r.db.Connection().Find(&users).Error
	return users, err
}

func (r *UserRepository) Exists(user *User) (bool, error) {
	var count int64
	err := r.db.Connection().Model(&User{}).Where("email = ?", user.Email).Count(
		&count,
	).Error
	return count > 0, err
}

func (r *UserRepository) Delete(user *User) error {
	conn := r.db.Connection()
	return conn.Where("email = ?", user.Email).Delete(user).Error
}
