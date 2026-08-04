package models

import "gorm.io/gorm"

type Booking struct {
	gorm.Model
	BookingCode string `json:"bokingCode"`
	Phone       string `json:"phone"`
	
	UserID int    `json:"userId"`
	User   User   `gorm:"foreignkey:UserID" json:"user"`
	
	EventID int    `json:"eventId"`
	Event   Event  `gorm:"foreignkey:EventID" json:"event"`
	
}
