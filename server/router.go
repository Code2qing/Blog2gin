package server

import (
	"Blog2Gin/server/controllers"
	"github.com/gin-gonic/gin"
)

func InitRouter(r *gin.Engine) {
	r.GET("/", controllers.BlogIndex)
	r.Static("/static", "static")
}
