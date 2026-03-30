package controllers

import (
	"github.com/gin-gonic/gin"
)

func quotesRoutes(r gin.RouterGroup) {
	rgroup := r.Group("/quotes")
	rgroup.GET("/", GetQuotes)
	rgroup.GET("/:id", GetQuoteByID)
	rgroup.POST("/", CreateQuote)
	rgroup.PUT("/:id", UpdateQuote)
	rgroup.DELETE("/:id", DeleteQuote)
	rgroupt := rgroup.Group("/archive")
	rgroupt.DELETE("/trash", TrashQuotes)
	rgroupt.DELETE("/restore", RestoreQuote)
}

func GetQuotes(c *gin.Context) {}

func GetQuoteByID(c *gin.Context) {}

func CreateQuote(c *gin.Context) {}

func UpdateQuote(c *gin.Context) {}

func DeleteQuote(c *gin.Context) {}

func TrashQuotes(c *gin.Context) {}

func RestoreQuote(c *gin.Context) {}