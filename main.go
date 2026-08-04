package main

import (
	"log"

	"example.com/event/config"
	"example.com/event/controllers"
	"example.com/event/middlewares"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/gin-contrib/cors"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
	}
	config.ConnectDB()
	server := gin.Default()

	server.Use(cors.Default())

	//routes and handlers 
	api := server.Group("/api")
	{
		api.GET("/events", controllers.GetEvents)
		api.GET("/events/:id", controllers.GetEventById)

		api.POST("/auth/register", controllers.RegisterUser)
		api.POST("/auth/login", controllers.LoginUser)

		protected :=api.Group("/")
		protected.Use(middlewares.RequireAuth)
		{
			protected.GET("/events/user", controllers.GetEventsByUser)
			protected.GET("/auth/me", controllers.GetCurrentUser)
			protected.POST("/events", controllers.CreateEvent)
			protected.PUT("/events/:id", controllers.UpdateEvent)
			protected.DELETE("/events/:id", controllers.DeleteEvent)

			protected.POST("/booking", controllers.CreateBookingEvent)
			protected.GET("/booking/user", controllers.GetBookingbyUser)
			protected.DELETE("/booking/:id", controllers.DeleteBooking)
		}

	}
	server.Run(":8080")
}
