package main

import (
	"project-devis-quote/controllers"

	"github.com/gin-gonic/gin"
)

func setupRouter() *gin.Engine {
	r := gin.Default()
	quote := r.Group("/api/quote")

	quote.GET("/", controllers.ListQuotes)
	quote.POST("/", controllers.CreateQuote)
	quoteId := quote.Group("/:id")
	{
		quoteId.GET("/", controllers.GetQuote)
		quoteId.PUT("/", controllers.UpdateQuote)
	}
	return r
}

func main() {
	r := setupRouter()

	r.Run(":8080")
}