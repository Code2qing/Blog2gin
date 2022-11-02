package model

import (
	"Blog2Gin/conf"
	"database/sql"
	"reflect"
	"time"
)

type Category struct {
	ID        uint        `json:"id" gorm:"primaryKey"`
	Name      string      `json:"name"`
	BlogPosts []*BlogPost `json:"blog_posts"`
}

type CategoryCount struct {
	Category
	CategoryCount int `gorm:"-:migration"`
}

type Archive struct {
	Year  int
	Month int
	Count int
}

type Tag struct {
	TagID     uint        `json:"id" gorm:"primaryKey;column:id"`
	TagName   string      `json:"name" gorm:"column:name"`
	BlogPosts []*BlogPost `gorm:"many2many:blog_post_tags;joinForeignKey:TagID;foreignKey:TagID;joinReferences:PostID;references:ID"`
}

type TagCount struct {
	Tag
	BlogPostTags []*BlogPostTags `gorm:"-;foreignKey:TagID"`
	TagPostCount int             `gorm:"-:migration"`
}

type BlogPostTags struct {
	ID     uint
	TagID  uint
	PostID uint
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
	Tags         []Tag     `gorm:"many2many:blog_post_tags;joinForeignKey:PostID;foreignKey:ID;joinReferences:TagID;references:TagID"`
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

func GetCategoryCount() (int, error) {
	var count int64
	if err := conf.DB.Model(&Category{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return int(count), nil
}

func GetTagCount() (int, error) {
	var count int64
	if err := conf.DB.Model(&Tag{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return int(count), nil
}

func structAssign(binding interface{}, value interface{}) {
	bVal := reflect.ValueOf(binding).Elem() //获取reflect.Type类型
	vVal := reflect.ValueOf(value).Elem()   //获取reflect.Type类型
	vTypeOfT := vVal.Type()
	for i := 0; i < vVal.NumField(); i++ {
		// 在要修改的结构体中查询有数据结构体中相同属性的字段，有则修改其值
		name := vTypeOfT.Field(i).Name
		if ok := bVal.FieldByName(name).IsValid(); ok {
			bVal.FieldByName(name).Set(reflect.ValueOf(vVal.Field(i).Interface()))
		}
	}
}

func GetPostDetailWithTagCate(postID int) (blogPost *BlogPost, err error) {
	var tags []Tag
	// 联表查询
	//err = conf.DB.Order("blog_post.id desc").Joins("Category").Where("blog_post.id = ?", postID).Preload("Tags").Find(&blogPost).Error
	rows, err := conf.DB.Raw("SELECT `blog_post`.`id`,`blog_post`.`title`,`blog_post`.`body`,`blog_post`"+
		".`body_md`,`blog_post`.`excerpt`,`blog_post`.`views`,`blog_post`.`category_id`,`blog_post`.`created_time`,"+
		"`blog_post`.`modified_time`,`Category`.`id` AS `Category__id`,`Category`.`name` AS `Category__name`,"+
		"`Tags`.`id` AS `TagID`,`Tags`.`name` AS `TagName` FROM `blog_post` "+
		"LEFT JOIN `blog_category` `Category` ON `blog_post`.`category_id`=`Category`.`id` "+
		"LEFT JOIN `blog_post_tags` ON `blog_post`.`id`=`blog_post_tags`.`post_id` "+
		"LEFT JOIN `blog_tag` `Tags` ON `blog_post_tags`.`tag_id`=`Tags`.`id` "+
		"WHERE blog_post.id=?", postID).Rows()
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		err = rows.Close()
		if err != nil {
			blogPost = nil
		}
	}(rows)
	for rows.Next() {
		err = conf.DB.ScanRows(rows, &blogPost)
		if err != nil {
			return nil, err
		}
		err = conf.DB.ScanRows(rows, &tags)
		if err != nil {
			return nil, err
		}
	}
	blogPost.Tags = tags
	return blogPost, nil
}

func GroupArchive() ([]*Archive, error) {
	var Archives []*Archive
	err := conf.DB.Model(&BlogPost{}).Select("YEAR(created_time) as `year`, MONTH(created_time) as `month`, count(*) as `count`").Group("`year` desc, `month` desc").Find(&Archives).Error
	if err != nil {
		return nil, err
	}
	return Archives, nil
}

func ArchivePosts(year int, month int) ([]*BlogPost, error) {
	var blogPosts []*BlogPost
	err := conf.DB.Where("YEAR(created_time)=? AND MONTH(created_time)=?", year, month).Joins("Category").Order("id desc").Find(&blogPosts).Error
	if err != nil {
		return nil, err
	}
	return blogPosts, nil
}

func TagList() ([]*TagCount, error) {
	var tags []*TagCount
	err := conf.DB.Model(&TagCount{}).Select("blog_tag.id AS TagID, blog_tag.`name` AS TagName, " +
		"count(BlogPostTag.tag_id) as TagPostCount").Joins("LEFT JOIN `blog_post_tags` `BlogPostTag` " +
		"ON `blog_tag`.`id` = `BlogPostTag`.`tag_id`").Group("blog_tag.id").Find(&tags).Error
	//err := conf.DB.Select("blog_tag.id AS TagID, blog_tag.`name` AS TagName, count(BlogPostTags.tag_id) as TagPostCount").Joins("BlogPostTags").Group("blog_tag.id").Find(&tags).Error
	if err != nil {
		return nil, err
	}
	return tags, nil
}

func TagPosts(tagID int) ([]*BlogPost, error) {
	var blogPosts []*BlogPost
	err := conf.DB.Model(&Tag{TagID: uint(tagID)}).Joins("Category").Order("id desc").Association("BlogPosts").Find(&blogPosts)
	if err != nil {
		return nil, err
	}
	return blogPosts, nil
}

func CategoryList() ([]*CategoryCount, error) {
	var categories []*CategoryCount
	err := conf.DB.Select("blog_category.id as ID, blog_category.name as Name, count(BlogPosts.category_id) as" +
		" CategoryCount").Joins("LEFT JOIN `blog_post` `BlogPosts` ON `blog_category`.`id` = " +
		"`BlogPosts`.`category_id` ").Group("blog_category.id").Find(&categories).Error
	if err != nil {
		return nil, err
	}
	return categories, nil
}

func CategoryPosts(categoryID int) ([]*BlogPost, error) {
	var blogPosts []*BlogPost
	err := conf.DB.Model(&Category{ID: uint(categoryID)}).Joins("Category").Order("id desc").Association("BlogPosts").Find(&blogPosts)
	if err != nil {
		return nil, err
	}
	return blogPosts, nil
}

//func TagPostCount([]*Tag) []*Tag {
//	var count int64
//	rows, err := conf.DB.Raw("select count(*) as tag_post_count, tag_id from blog_post_tags  GROUP BY tag_id HAVING tag_id in ")
//	return int(count)
//}

//func GetPostsByCategory(categoryID int) ([]BlogPost, error) {
//	var BlogPosts []BlogPost
//	err := conf.DB.Order("`blog_post`.`id` desc").Joins("Category").Where("Category.id = ?", categoryID).Find(&BlogPosts).Error
//	if err != nil {
//		return nil, err
//	}
//	return BlogPosts, nil
//}
