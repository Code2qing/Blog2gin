package forms

type IndexParam struct {
	PageNum int `form:"page,default=1"`
}

type DetailParam struct {
	PostID int `uri:"postID" binding:"required"`
}

type ArchiveParam struct {
	Year  int `uri:"year" binding:"required"`
	Month int `uri:"month" binding:"required"`
}

type TagParam struct {
	TagID int `uri:"tag_id" binding:"required"`
}
type CategoryParam struct {
	CategoryID int `uri:"category_id" binding:"required"`
}
