package controllers

import (
	"net/http"
	"os"
	"time"

	"example.com/event/config"
	"example.com/event/models"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AutoInputRegister struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type AutoInputLogin struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func RegisterUser(c *gin.Context) {
	var input AutoInputRegister

	//validation
	err := c.ShouldBindJSON(&input)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	//hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to hash password"})
		return
	}

	//simpan ke DB
	user := models.User{
		Name:     input.Name,
		Email:    input.Email,
		Password: string(hashedPassword),
	}

	userCreated := config.DB.Create(&user)
	if userCreated.Error != nil {
		c.JSON(500, gin.H{"error": "failed to create user"})
		return
	}
	c.JSON(201, gin.H{"message": "user created successfully",
		"user": gin.H{
			"name": user.Name,
			"email": user.Email,
			"id": user.ID,
			"created_at": user.CreatedAt,
			"updated_at": user.UpdatedAt,
		}})
}

func LoginUser(c *gin.Context){
	var input AutoInputLogin
	if err := c.ShouldBindJSON(&input); err != nil{
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	userData:=config.DB.Where("email = ?", input.Email).First(&user)

	if userData.Error != nil {
		c.JSON(401, gin.H{"error": "user not found"})
		return
	}

	errMatch:=bcrypt.CompareHashAndPassword([]byte(user.Password),[]byte(input.Password))
	
    if errMatch!=nil{
        c.JSON(401, gin.H{"error": "password salah"})
        return
    }

	//buat token 

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err:= token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err!= nil {
		c.JSON(500, gin.H{"error": "failed to generate token"})
		return
	}
	c.JSON(200, gin.H{
		"message": "Login successfully",
		"token": tokenString,
		"user": gin.H{
			"name": user.Name,
			"email": user.Email,
			"id": user.ID,
			"created_at": user.CreatedAt,
			"updated_at": user.UpdatedAt,
		},
	})
}

func GetCurrentUser(c *gin.Context){

    //ambil dari 
	userId, exists := c.Get("userID")
	if !exists {
		c.JSON(401, gin.H{"error": "user not authenticated"})
		return
	}
	
	var user models.User
	userData := config.DB.Select("id", "name", "email").First(&user, userId).Error
	if userData != nil {
		c.JSON(500, gin.H{"error": "failed to get user"})
		return
	}

	c.JSON(http.StatusOK,
		gin.H{
			"user": gin.H{
				"id": user.ID,
				"name": user.Name,
				"email": user.Email,
				"events": user.Events,
			},
		},
	)
}

