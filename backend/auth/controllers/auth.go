package controllers

import (
	"database/sql"
	"log"
	"net/http"
	"project-devis-auth/services"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	_ "github.com/lib/pq"
)

type RegisterInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type LoginInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func init() {
	// Initialize database connection here if needed
	connStr := "user=postgres password=yourpassword dbname=yourdb sslmode=disable"
	 db, err := sql.Open("postgres", connStr)
	 if err != nil {
	 	log.Fatal(err)
	 }
	 defer db.Close()
	
}


func Register(c *gin.Context) {
	var input RegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// Here you would typically hash the password and store the user in the database
	db := services.ConnectDB()
	
	db.Exec("SELECT * FROM users WHERE email = $1", input.Email)
	if _, err := services.HashPassword(input.Password); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "registration successful!"})
}

func Login(c *gin.Context) {
	var input LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := services.ConnectDB()

	var password string
	err := db.QueryRow("SELECT password FROM Auth where email = $1", input.Email).Scan(&password)

	if err == nil && services.VerifyPassword(input.Password, password) {
		token, err := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.MapClaims{
			"email": input.Email,
			"exp":   jwt.NewNumericDate(time.Now().Add(time.Hour)),
		}).SignedString([]byte(services.APPKey.GetValue()))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"token": token})
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
	}
}
