package model

import "github.com/dgrijalva/jwt-go"

type User struct {
	Account  string `form:"account" json:"account" binding:"required"`
	Username string `form:"username" json:"username" binding:"required"`
	Password string `form:"password" json:"password" binding:"required"`
}

type UserLogin struct {
	Account  string `form:"account" json:"account" binding:"required"`
	Username string `form:"username" json:"username"`
	Password string `form:"password" json:"password" binding:"required"`
}

type MyClaims struct {
	Account string `json:"account"`
	jwt.StandardClaims
}
