package model

import (
	"Blog2Gin/conf"
	"time"
)

type Category struct {
	ID   uint   `json:"id" gorm:"primaryKey"`
	Name string `json:"name"`
	//BlogPosts []*BlogPost `json:"blog_posts"`
}

type Tag struct {
	ID        uint       `json:"id" gorm:"primaryKey"`
	Name      string     `json:"name"`
	BlogPosts []BlogPost `gorm:"many2many:blog_post_tags"`
}

type BlogPost struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	Title        string    `json:"title"`
	Body         string    `json:"body"`
	BodyMd       string    `json:"body_md"`
	Excerpt      string    `json:"excerpt"`
	Views        uint      `json:"views"`
	CategoryID   uint      `json:"category_id"`
	Category     Category  `gorm:"foreignKey:CategoryID"`
	Tags         []Tag     `gorm:"many2many:blog_post_tags;joinForeignKey:PostID"`
	CreatedTime  time.Time `json:"created_time" gorm:"autoCreateTime:true"`
	ModifiedTime time.Time `json:"modified_time" gorm:"autoUpdateTime:true"`
}

func (Category) TableName() string {
	return "blog_category"
}

func (Tag) TableName() string {
	return "blog_tag"
}

func (BlogPost) TableName() string {
	return "blog_post"
}

func GetPostsByPage(pageNum int, pageSize int) ([]*BlogPost, error) {
	var BlogPosts []*BlogPost
	err := conf.DB.Order("id desc").Limit(pageSize).Offset(pageSize * (pageNum - 1)).Joins("Category").Find(&BlogPosts).Error
	if err != nil {
		return nil, err
	}
	return BlogPosts, nil
}

func GetPostCount() (int, error) {
	var totalCount int64
	if err := conf.DB.Model(&BlogPost{}).Count(&totalCount).Error; err != nil {
		return 0, err
	}
	return int(totalCount), nil
}

func GetPostsByCategory(categoryID int) ([]BlogPost, error) {
	var BlogPosts []BlogPost
	err := conf.DB.Order("`blog_post`.`id` desc").Joins("Category").Where("Category.id = ?", categoryID).Find(&BlogPosts).Error
	if err != nil {
		return nil, err
	}
	return BlogPosts, nil
}
