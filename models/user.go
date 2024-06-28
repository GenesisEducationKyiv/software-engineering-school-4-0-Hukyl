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
