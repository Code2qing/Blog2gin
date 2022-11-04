package common

import (
	"Blog2Gin/model"
	"github.com/gin-gonic/gin"
	"time"
)

type BaseCtxData struct {
	MenuHome      bool
	MenuArchive   bool
	MenuTag       bool
	NowYear       int
	PostCount     int
	CategoryCount int
	TagCount      int
	GinCtx        *gin.Context
	Title         string
}

type IndexContextData struct {
	Left         []int
	Right        []int
	LeftHasMore  bool
	RightHasMore bool
	First        bool
	Last         bool
	BlogPosts    []*model.BlogPost
	Ispaginated  bool
	PageNum      int
	TotalPageNum int
	BaseCtxData
}

type DetailCtxData struct {
	BaseCtxData
	BlogPost *model.BlogPost
}
type ArchivesCtxData struct {
	BaseCtxData
	Archives []*model.Archive
}

type TagCtxData struct {
	BaseCtxData
	Tags []*model.TagCount
}

type CategoryCtxData struct {
	BaseCtxData
	Categories []*model.CategoryCount
}

func DefaultBaseCtxData() *BaseCtxData {
	nowYear, _, _ := time.Now().Date()
	return &BaseCtxData{NowYear: nowYear}
}

func DefaultIdxCtxData() *IndexContextData {
	return &IndexContextData{BaseCtxData: *DefaultBaseCtxData(), PageNum: 1, TotalPageNum: 1}
}
func DefaultDetailCtxData() *DetailCtxData {
	return &DetailCtxData{BaseCtxData: *DefaultBaseCtxData()}
}
func DefaultArchivesCtxData() *ArchivesCtxData {
	return &ArchivesCtxData{BaseCtxData: *DefaultBaseCtxData()}
}
func DefaultTagsCtxData() *TagCtxData {
	return &TagCtxData{BaseCtxData: *DefaultBaseCtxData()}
}
func DefaultCategoriesData() *CategoryCtxData {
	return &CategoryCtxData{BaseCtxData: *DefaultBaseCtxData()}
}
