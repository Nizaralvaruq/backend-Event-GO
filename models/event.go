package models

import (
	
	"time"
	"gorm.io/gorm"
)

type Event struct {
	gorm.Model
	Name        string `json:"name" binding:"required"`
	Description string `json:"description" binding:"required"`
	Location    string `json:"location" binding:"required"`
	Image       string `json:"image"`
	ImageID     string `json:"imageId"`
	UserID      uint   `json:"userId"`
	User        User   `gorm:"foreignkey:UserID" json:"user"`
	Datetime    time.Time `json:"datetime" binding:"required"`
	Booking     []Booking `gorm:"foreignkey:EventID" json:"booking"`
}
