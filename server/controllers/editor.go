package controllers

import (
	"Blog2Gin/conf"
	"Blog2Gin/model"
	"github.com/gin-gonic/gin"
	"net/http"
)

type editorParam struct {
	PostID int `uri:"postID" binding:"required"`
}

type updateParam struct {
	Passwd  string `json:"passwd" binding:"required"`
	PostID  int    `json:"post_id" binding:"required"`
	Md      string `json:"md" binding:"required"`
	Excerpt string `json:"excerpt" binding:"required"`
}

func Editor(c *gin.Context) {
	var args editorParam
	if err := c.BindUri(&args); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	postID := args.PostID
	var blogPost model.BlogPost
	if err := conf.DB.Where(&model.BlogPost{ID: uint(postID)}).First(&blogPost).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.HTML(200, "edit_post.html", gin.H{"BlogPost": blogPost, "PostID": postID})
}

func UpdateMd(c *gin.Context) {
	var args updateParam
	if err := c.BindJSON(&args); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": err.Error()})
		return
	}
	passwd := args.Passwd
	md := args.Md
	postID := args.PostID
	excerpt := args.Excerpt
	if passwd != conf.Conf.UpdatePasswd {
		c.AbortWithStatusJSON(200, gin.H{"msg": "forbidden!"})
		return
	}
	err := conf.DB.Model(&model.BlogPost{ID: uint(postID)}).Updates(map[string]any{"Body": md, "Excerpt": excerpt}).Error
	if err != nil {
		c.AbortWithStatusJSON(200, gin.H{"msg": err.Error()})
	}
	c.JSON(http.StatusOK, gin.H{"msg": "ok"})
}
