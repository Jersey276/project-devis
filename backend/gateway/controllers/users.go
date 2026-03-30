package controllers

import (
	"github.com/gin-gonic/gin"
)

func UserRoutes(r *gin.RouterGroup) {
	ruser := r.Group("me")
	ruser.GET("/", GetInfo)
	ruser.PUT("/", UpdateInfo)
	
	ruseraddr := ruser.Group("/address")
	ruseraddr.GET("/", ListAddress)
	ruseraddr.POST("/", AddAddress)
	ruseraddr.PUT("/:id", UpdateAddress)
	ruseraddr.DELETE("/:id", DeleteAddress)	
}

func GetInfo(c *gin.Context) {}

func UpdateInfo(c *gin.Context) {}

func ListAddress(c *gin.Context) {}

func AddAddress(c *gin.Context) {}

func UpdateAddress(c *gin.Context) {}

func DeleteAddress(c *gin.Context) {}