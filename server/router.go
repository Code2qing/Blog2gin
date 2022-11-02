package server

import (
	"Blog2Gin/server/controllers"
	"github.com/gin-gonic/gin"
)

func InitRouter(r *gin.Engine) {
	r.GET("/", controllers.BlogIndex)
	r.GET("/post/:postID", controllers.BlogDetail)
	r.GET("/archives", controllers.Archives)
	r.GET("/archives/:year/:month", controllers.ArchivePosts)
	r.GET("/tags", controllers.Tags)
	r.GET("/tags/:tag_id", controllers.TagPosts)
	r.GET("/categories", controllers.Categories)
	r.GET("/category/:category_id", controllers.CategoryPosts)
	r.Static("/static", "static")
}
