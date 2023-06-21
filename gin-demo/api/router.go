package api

import (
	"gin-demo/api/middleware"
	"github.com/gin-gonic/gin"
)

func InitRouter() {
	r := gin.Default()

	r.Use(middleware.CORS())
	r.Use(middleware.JWTAuthMiddleware())
	rGroup := r.Group("/login")
	{
		rGroup.POST("", login)
	}
	//r.POST("/username", GetAccountFromToken)
	r.POST("/register", register) // 注册
	r.POST("/raisequestion", RaiseQuestion)
	r.POST("/question", ShowQuestions)
	r.POST("/givecomment", GiveComments)
	r.POST("/myquestion", ShowMyQuestion)
	r.POST("/deletemycomment", DeleteMyComments)
	r.POST("/deletemyquestion", DeleteMyQuestion)
	r.POST("/updatemyquestion", UpdateMyQuestion)

	r.Run(":8080")
}
