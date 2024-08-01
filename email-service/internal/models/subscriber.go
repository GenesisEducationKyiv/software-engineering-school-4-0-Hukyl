package models

import "fmt"

type Subscriber struct {
	ID      uint   `gorm:"primaryKey"`
	Email   string `gorm:"unique"`
	Created int64  `gorm:"autoCreateTime"` // Use unix seconds as creating time
}

func (s Subscriber) String() string {
	return fmt.Sprintf("Subscriber<%d, %s>", s.ID, s.Email)
}

func (Subscriber) TableName() string {
	return "mail_subscribers"
}
