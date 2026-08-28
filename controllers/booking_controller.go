package controllers

import (
	"fmt"
	"net/http"
	"time"

	"example.com/event/config"
	"example.com/event/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type BokingInput struct {
    EventID int    `json:"eventId" binding:"required"`  
    Phone   string `json:"phone" binding:"required"`    
}


func CreateBookingEvent(c *gin.Context){
	userID, _ := c.Get("userID")

	var input BokingInput
	var booking models.Booking

	if errValidation := c.ShouldBindJSON(&input); errValidation != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": errValidation.Error(),
		})
		return
	}

	//kondisi jika user pernah boking event ID
	Booking := config.DB.Where("user_id = ? AND event_id = ?", userID, input.EventID).First(&booking)
	if Booking.Error == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Anda sudah pernah booking event ini",
		})
		return
	}

	//generate booking code
    CodeBoking := fmt.Sprintf("BK-%sE%dU%d", 
    time.Now().Format("20060102"), input.EventID, userID.(uint)) 

    //simpan ke DB
    bookingData := models.Booking{
    EventID:     input.EventID,
    UserID:      int(userID.(uint)), 
    BookingCode: CodeBoking,
    Phone:       input.Phone,
}

	errCreateBooking := config.DB.Create(&bookingData).Error
	if errCreateBooking != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "gagal membuat booking",
			"error":   errCreateBooking.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "booking created successfully",
		"booking": bookingData,
	})
}

func GetBookingbyUser(c *gin.Context){
	var booking []models.Booking
	userID, _ := c.Get("userID")

	errBookingData := config.DB.Preload("Event").Preload("Event.User").Preload("User",func(db *gorm.DB) *gorm.DB{
		return db.Select("id", "name", "email")
	}).Where("user_id = ?", userID).Find(&booking).Error

	if errBookingData != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to retrieve booking / booking not found",
			"error": errBookingData.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "bookings retrieved successfully",
		"bookings": booking,
	})
}

func DeleteBooking(c *gin.Context){
	userID, _ := c.Get("userID")
	var booking models.Booking
	
	ParamsId := c.Param("id")
	bookingData := config.DB.First(&booking,ParamsId).Error
	if bookingData != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "failed to retrieve booking / booking not found",
			"error":   bookingData.Error(),
		})
		return
	}

	if booking.UserID != int(userID.(uint)){
		c.JSON(http.StatusForbidden, gin.H{
			"message": "forbidden / kamu tidak bisa menghapus booking orang lain",
		})
		return
	}

	errDeleteBooking := config.DB.Unscoped().Delete(&booking).Error
	if errDeleteBooking != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "gagal menghapus booking",
			"error":   errDeleteBooking.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "booking deleted successfully",
	})
}