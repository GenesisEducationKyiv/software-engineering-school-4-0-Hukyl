package models

import (
	"fmt"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Email string `json:"email"`
}

func (u User) String() string {
	return fmt.Sprintf("User<%d, %#v>", u.ID, u.Email)
}

func UserExists(db *DB, email string) (bool, error) {
	var exists bool
	err := db.Connection().Model(&User{}).
		Select("count(*) > 0").
		Where("email = ?", email).
		Find(&exists).
		Error
	return exists, err
}
