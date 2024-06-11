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

func (r *UserRepository) FindByEmail(email string) (*User, error) {
	var user User
	err := r.db.Connection().Where("email = ?", email).First(&user).Error
	return &user, err
}

func (r *UserRepository) FindByID(id uint) (*User, error) {
	var user User
	err := r.db.Connection().First(&user, id).Error
	return &user, err
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

func (r *UserRepository) Update(user *User) error {
	return r.db.Connection().Save(user).Error
}

func (r *UserRepository) Delete(user *User) error {
	return r.db.Connection().Delete(user).Error
}
