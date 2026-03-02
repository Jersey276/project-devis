package main

import (
	"project-devis-users/controllers"

	"github.com/gin-gonic/gin"
)

func setupRouter() *gin.Engine {
	r := gin.Default()
	api := r.Group("/api")
	users := api.Group("/users")
	
	

	users.GET("/", controllers.GetMe)

	// Address routes
	address := users.Group("/address")
	address.GET("/", controllers.ListAddresses)
	address.POST("/", controllers.CreateAddress)
	addressId := address.Group("/:id")
	addressId.GET("/", controllers.GetAddress)
	addressId.PUT("/", controllers.UpdateAddress)
	addressId.DELETE("/", controllers.DeleteAddress)

	// Customer routes
	customer := users.Group("/customer")
	customer.GET("/", controllers.ListCustomer)
	customer.POST("/", controllers.CreateCustomer)
	customerId := customer.Group("/:id")
	customerId.GET("/", controllers.GetCustomer)
	customerId.DELETE("/", controllers.DeleteCustomer)

	// Country and Country Group routes
	country := api.Group("/country")
	country.GET("/", controllers.ListCountries)
	countryGroup := country.Group("/:id")
	countryGroup.GET("/", controllers.GetCountry)

	return r
}

func main() {
	r := setupRouter()

	r.Run(":8080")
}