package domain

import (
	"time"

	"github.com/google/uuid"
)

type Subscription struct {
	ID          uint       `gorm:"primaryKey;autoIncrement"`
	ServiceName string     `gorm:"column:service_name;not null"`
	Price       int        `gorm:"column:price;not null"`
	UserID      uuid.UUID  `gorm:"column:user_id;type:uuid;not null"`
	StartDate   time.Time  `gorm:"column:start_date;not null"`
	EndDate     *time.Time `gorm:"column:end_date"`
}

func (Subscription) TableName() string {
	return "subscriptions"
}
