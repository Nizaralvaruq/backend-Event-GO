package controllers

import (
	"context"
	"math"
	"net/http"
	"os"
	"strconv"
	"time"

	"example.com/event/config"
	"example.com/event/models"
	"github.com/gin-gonic/gin"
	"github.com/imagekit-developer/imagekit-go/v2"
	"github.com/imagekit-developer/imagekit-go/v2/option"
	"gorm.io/gorm"
)

func initImageKit() imagekit.Client {
	client := imagekit.NewClient(
		option.WithPrivateKey(os.Getenv("IMAGEKIT_PRIVATE_KEY")),
	)
	return client
}

func CreateEvent(c *gin.Context){
	userID := c.MustGet("userID").(uint)

	file, header, err := c.Request.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	defer file.Close()

	//1.upload file ke imagekit
	fileName := header.Filename
	ik := initImageKit()

	uploadRes, errUpload := ik.Files.Upload(context.Background(), imagekit.FileUploadParams{
		File:     file,
		FileName: fileName,
	})
	if errUpload != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "upload gagal",
			"details": errUpload.Error(),
		})
		return
	}

	
	parsedTime, errTime := time.Parse(time.RFC3339, c.PostForm("datetime"))
	if errTime != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "format datetime salah, gunakan format RFC3339",
		})
		return
	}
	
	//simpan ke DB
	event := models.Event{
		Name:        c.PostForm("name"),
		Description: c.PostForm("description"),
		Location:    c.PostForm("location"),
		Datetime:    parsedTime,
		Image:       uploadRes.URL,
		ImageID:     uploadRes.FileID,
		UserID:      userID,
	}
	
	config.DB.Create(&event)
	c.JSON(http.StatusOK, gin.H{
		"message": "Event created successfully",
		"event":   event,
	})
}

func GetEvents(c *gin.Context){
	var events []models.Event
	
	//1.inisiasi dasar
	query := config.DB.Model(&models.Event{})
	
	//2.tangkap fungsi filter by query
	search :=c.Query("search")
	if search != ""{
		query = query.Where("name LIKE ? OR description LIKE ?", "%"+search+"%", "%"+search+"%")
	}
	
	//3.pagination(total data sebelum di limit)
	var totalRows int64
	query.Count(&totalRows)

	//4. tangkap parameter query dan masukan nilai default
	pageStr :=c.DefaultQuery("page","1")
	limitStr :=c.DefaultQuery("limit","10")

	//ubah ke integer
	page, errPage := strconv.Atoi(pageStr)
	if errPage != nil || page < 1{
		page = 1	
	}

	limit, errLimit := strconv.Atoi(limitStr)
	if errLimit != nil || limit < 1{
		limit = 10
	}

	// 5.hitung offset
	offset := (page - 1) * limit
	
	//6.hitung data perhalaman
	totalPages := int(math.Ceil(float64(totalRows) / float64(limit)))
	
	//7.eksekesi fitur semua di atas
	if errEvent := query.Preload("User", func(db *gorm.DB) *gorm.DB{
		return db.Select("id", "name", "email")
	}).Offset(offset).Limit(limit).Find(&events).Error; errEvent != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "gagal memangil data event",
			"error":   errEvent.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "events retrieved successfully",
		"total_rows": totalRows,
		"total_pages": totalPages,
		"page": page,
		"limit": limit,
		"events": events,
	})
}

func GetEventById(context *gin.Context){
	var event models.Event
	paramsId := context.Param("id")

	var eventData = config.DB.Preload("User",func (db *gorm.DB) *gorm.DB {
		return db.Select("id","name","email")
	}).Preload("Booking").Preload("Booking.User",func(db *gorm.DB) *gorm.DB{
		return db.Select("id", "name", "email")
	})
	err := eventData.First(&event, paramsId).Error

	if err != nil {
		context.JSON(http.StatusNotFound, gin.H{ // Menggunakan 404 Not Found
			"message": "failed to retrieve event / event not found",
			"error":   err.Error(),
		})
		return
	}

	context.JSON(http.StatusOK, gin.H{
		"message": "event retrieved successfully",
		"event":   event,
	})
}

//get event by id yang membuat event (user)
func GetEventsByUser(c *gin.Context){
	var events []models.Event
	userID := c.MustGet("userID").(uint)

	errEvent := config.DB.Where("user_id = ?", userID).Preload("User",func(db *gorm.DB) *gorm.DB{
		return db.Select("id", "name", "email")
	}).Find(&events).Error

	if errEvent != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "failed to retrieve event / event not found",
			"error":   errEvent.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "events retrieved successfully",
		"events": events,
	})
}

func UpdateEvent(c *gin.Context){
	userID := c.MustGet("userID").(uint)
	var event models.Event
	paramsId := c.Param("id")

	// 1. Cari data event yang lama di database terlebih dahulu
	err := config.DB.First(&event, paramsId).Error
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "failed to retrieve event / event not found",
			"error":   err.Error(),
		})
		return
	}

	if event.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{
			"message": "forbidden / kamu tidak bisa mengupdate event orang lain",
		})
		return
	}

	file, header, err := c.Request.FormFile("image")
	if err == nil {
		// Ada gambar baru dikirim, proses upload
		defer file.Close()

		fileName := header.Filename
		ik := initImageKit()

		uploadRes, errUpload := ik.Files.Upload(context.Background(), imagekit.FileUploadParams{
			File:     file,
			FileName: fileName,
		})
		if errUpload != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "upload gambar gagal",
				"details": errUpload.Error(),
			})
			return
		}

		// Hapus gambar lama dari ImageKit
		if event.ImageID != "" {
			ik.Files.Delete(context.Background(), event.ImageID)
		}

		// Set gambar baru
		event.Image = uploadRes.URL
		event.ImageID = uploadRes.FileID
	}

    if name := c.PostForm("name"); name != ""{ 
		event.Name = name
	}

	if description := c.PostForm("description"); description != ""{
		event.Description = description
	}

	if location := c.PostForm("location"); location != ""{
		event.Location = location
	}

	if dateTimeStr := c.PostForm("datetime"); dateTimeStr != ""{
		parsedTime, errParse := time.Parse(time.RFC3339, dateTimeStr)
		if errParse != nil{
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "format datetime salah, gunakan format RFC3339",
			})
			return
		}
		event.Datetime = parsedTime
	}

	config.DB.Save(&event)

	c.JSON(http.StatusOK, gin.H{
		"message": "Event updated successfully",
		"event":   event,
	})
}


func DeleteEvent(c *gin.Context){
	
	userID := c.MustGet("userID").(uint)
	var event models.Event
	paramsId := c.Param("id")

	err := config.DB.First(&event, paramsId).Error
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "failed to retrieve event / event not found",
			"error":   err.Error(),
		})
		return
	}

	if event.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "KAMU TIDAK BISA MENGHAPUS EVENT ORANG LAIN",
		})
		return
	}

	if event.ImageID != "" {
		ik := initImageKit()
		if err := ik.Files.Delete(context.Background(), event.ImageID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "failed to delete old image",
				"error":   err.Error(),
			})
			return
		}
	}
	config.DB.Unscoped().Delete(&event)
	c.JSON(http.StatusOK, gin.H{
		"message": "Event deleted successfully",
	})
}