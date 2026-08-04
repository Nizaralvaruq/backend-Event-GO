package middlewares

import (
	"fmt"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func RequireAuth(c *gin.Context){
	tokenString := c.GetHeader("Authorization")
	if tokenString == ""{
		c.JSON(401, gin.H{"error": "authorization token not provided"})
		c.Abort()
		return
	}

	if !strings.HasPrefix(tokenString, "Bearer "){
		c.JSON(401, gin.H{"error": "authorization token is not valid"})
		c.Abort()
		return
	}
	
	tokenString=strings.TrimPrefix(tokenString, "Bearer ")

	token, err:= jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	
	if err != nil {
		c.JSON(401, gin.H{"error": "authorization token is not valid"})
		c.Abort()
		return
	}
	
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		sub, ok := claims["sub"].(float64)
		if !ok {
			c.JSON(401, gin.H{"error": "authorization token is not valid"})
			c.Abort()
			return
		}
		c.Set("userID", uint(sub))
	} else {
		c.JSON(401, gin.H{"error": "authorization token is not valid"})
		c.Abort()
		return
	}
	c.Next()
}