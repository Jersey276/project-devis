package services

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type AuthClaims struct {
	Email            string `json:"email"`
	UserID           string `json:"user_id"`
	Role             string `json:"role"`
	AccountStatus    string `json:"account_status"`
	SubscriptionTier string `json:"subscription_tier"`
	jwt.RegisteredClaims
}

func GenerateAccessToken(email, userID, role, accountStatus, subscriptionTier string) (string, error) {
	key := []byte(APPKey.GetValue())

	claims := AuthClaims{
		Email:            email,
		UserID:           userID,
		Role:             role,
		AccountStatus:    accountStatus,
		SubscriptionTier: subscriptionTier,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(key)
}

func ValidateAccessToken(tokenStr string) (*AuthClaims, error) {
	key := []byte(APPKey.GetValue())

	token, err := jwt.ParseWithClaims(tokenStr, &AuthClaims{}, func(t *jwt.Token) (interface{}, error) {
		return key, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*AuthClaims)
	if !ok || !token.Valid {
		return nil, jwt.ErrTokenInvalidClaims
	}

	return claims, nil
}
