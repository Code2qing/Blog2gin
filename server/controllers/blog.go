package controllers

import (
	"Blog2Gin/conf"
	"Blog2Gin/model"
	"bytes"
	"fmt"
	toc "github.com/abhinav/goldmark-toc"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/util"
	"gitlab.com/golang-commonmark/markdown"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
	"net/http"
	"runtime"
	"strings"
	"time"
)

type indexParam struct {
	PageNum int `form:"page,default=1"`
}

type detailParam struct {
	PostID int `uri:"postID" binding:"required"`
}

type archiveParam struct {
	Year  int `uri:"year" binding:"required"`
	Month int `uri:"month" binding:"required"`
}

type tagParam struct {
	TagID int `uri:"tag_id" binding:"required"`
}
type categoryParam struct {
	CategoryID int `uri:"category_id" binding:"required"`
}

type baseCtxData struct {
	MenuHome      bool
	MenuArchive   bool
	MenuTag       bool
	NowYear       int
	PostCount     int
	CategoryCount int
	TagCount      int
	GinCtx        *gin.Context
}

type indexContextData struct {
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
	baseCtxData
}

type detailCtxData struct {
	baseCtxData
	BlogPost *model.BlogPost
}
type archivesCtxData struct {
	baseCtxData
	Archives []*model.Archive
}

type tagCtxData struct {
	baseCtxData
	Tags []*model.TagCount
}

type categoryCtxData struct {
	baseCtxData
	Categories []*model.CategoryCount
}

func BlogIndex(c *gin.Context) {
	var args indexParam
	if err := c.BindQuery(&args); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	pageNum := args.PageNum

	var g errgroup.Group

	//var blogPostChan = make(chan any)
	//var blogCtPostChan = make(chan any)
	//var totalCountChan = make(chan any)
	var dataChan = make(chan map[string]any, 1)

	//var errCH = make(chan error, 3)
	//go execModel(blogPostChan, errCH, func() (any, error) { return model.GetPostsByPage(pageNum, conf.IndexPageSize) })
	//go execModel(blogCtPostChan, errCH, func() (any, error) { return model.GetPostsByCategory(12) })
	//go execModel(totalCountChan, errCH, func() (any, error) { return model.GetPostCount() })
	execModel(&g, dataChan,
		func() (map[string]any, error) {
			ret, err := model.GetPostsByPage(pageNum, conf.IndexPageSize)
			return map[string]any{"blogPosts": ret}, err
		})
	//execModel(&g, dataChan,
	//	func() (map[string]any, error) {
	//		ret, err := model.GetPostsByCategory(1)
	//		return map[string]any{"blog_category_posts": ret}, err
	//	})
	tagCount, categoryCount, totalCount, err := baseCount(&g)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	//g.Go(execModel(
	//	dataChan,
	//	func() (map[string]any, error) {
	//		ret, err := model.GetPostsByPage(pageNum, conf.IndexPageSize)
	//		return map[string]any{"blog_posts": ret}, err
	//	}))
	//
	//g.Go(execModel(
	//	dataChan,
	//	func() (map[string]any, error) {
	//		ret, err := model.GetPostsByCategory(1)
	//		return map[string]any{"blog_category_posts": ret}, err
	//	}))
	//g.Go(execModel(
	//	dataChan,
	//	func() (map[string]any, error) {
	//		ret, err := model.GetPostCount()
	//		return map[string]any{"total_count": ret}, err
	//	}))
	//g.Go(func() error {
	//	ret, err := model.GetPostsByPage(pageNum, conf.IndexPageSize)
	//	dataChan <- map[string]any{"blog_posts": ret}
	//	return err
	//})
	//g.Go(func() error {
	//	ret, err := model.GetPostsByCategory(12)
	//	dataChan <- map[string]any{"blog_category_posts": ret}
	//	return err
	//})
	//g.Go(func() error {
	//	ret, err := model.GetPostCount()
	//	dataChan <- map[string]any{"total_count": ret}
	//	return err
	//})
	err = g.Wait()
	close(dataChan)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	dbData := map[string]any{}
	for data := range dataChan {
		for key := range data {
			dbData[key] = data[key]
		}
	}
	blogPostsV, ok := dbData["blogPosts"]
	if !ok {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "not total_count!"})
		return
	}
	blogPosts := blogPostsV.([]*model.BlogPost)
	start := time.Now()
	for _, post := range blogPosts {
		//markdown := goldmark.New(
		//	// 支持 GFM
		//	goldmark.WithExtensions(extension.GFM),
		//)
		//var buf bytes.Buffer
		//if err := markdown.Convert([]byte(post.Excerpt), &buf); err != nil {
		//	c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "markdown fail!"})
		//}
		//post.Excerpt = buf.String()

		//post.Excerpt = string(github_flavored_markdown.Markdown([]byte(post.Excerpt)))

		//post.Excerpt = string(blackfriday.Run([]byte(post.Excerpt), blackfriday.WithExtensions()))

		md := markdown.New(markdown.XHTMLOutput(true))
		post.Excerpt = md.RenderToString([]byte(post.Excerpt))

	}
	log.Infof("exec time is: %s", time.Since(start))
	indexCtx := DefaultIdxCtxData()
	indexCtx.TagCount = tagCount
	indexCtx.PostCount = totalCount
	indexCtx.BlogPosts = blogPosts
	indexCtx.CategoryCount = categoryCount
	indexCtx.MenuHome = true
	indexCtx.PageNum = pageNum
	indexCtx.GinCtx = c
	totalPages := (totalCount + conf.IndexPageSize - 1) / conf.IndexPageSize
	indexCtx.TotalPageNum = totalPages
	if totalPages > 1 {
		PaginationData(indexCtx, pageNum, totalPages)
		indexCtx.Ispaginated = true
	}

	c.HTML(http.StatusOK, "index.html", indexCtx)

	//blogPosts, err := model.GetPostsByPage(pageNum, conf.IndexPageSize)
	//go func() {
	//	blogPosts, err := model.GetPostsByPage(pageNum, conf.IndexPageSize)
	//	if err != nil {
	//		errCh <- err
	//	} else {
	//		blogPostCh <- blogPosts
	//	}
	//}()

	//if err != nil {
	//	c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	//	return
	//}

	//totalCount, err := model.GetPostCount()
	//if err != nil {
	//	c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	//	return
	//}

	//blogCtPosts, err := model.GetPostsByCategory(1)

	//go func() {
	//	blogCtPosts, err := model.GetPostsByCategory(1)
	//	if err != nil {
	//		errCh <- err
	//	} else {
	//		blogCtPostCH <- blogCtPosts
	//	}
	//}()
	//if err != nil {
	//	c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	//	return
	//}
	//var blogPosts any
	//var blogCtPosts any
	//var totalCount any
	//select {
	//case err := <-errCH:
	//	c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	//	return
	//case ret := <-blogPostChan:
	//	blogPosts = ret.([]model.BlogPost)
	//}
	//
	//select {
	//case err := <-errCH:
	//	c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	//	return
	//case blogCtPosts = <-blogCtPostChan:
	//}
	//
	//select {
	//case err := <-errCH:
	//	c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	//	return
	//case totalCount = <-totalCountChan:
	//}
	//c.JSON(http.StatusOK, resp)
}

//func execModel(ch chan any, errCH chan error, f func() (any, error)) {
//	ret, err := f()
//	if err != nil {
//		errCH <- err
//	} else {
//		ch <- ret
//	}
//}

func PaginationData(indexCtx *indexContextData, pageNum int, totalPages int) {
	pageRange := makeRange(1, totalPages)
	rightEnd := pageNum + 2
	if pageNum+2 > len(pageRange) {
		rightEnd = len(pageRange)
	}
	if pageNum == 1 {
		indexCtx.Right = pageRange[pageNum:rightEnd]
		if indexCtx.Right[len(indexCtx.Right)-1]+1 < totalPages {
			indexCtx.RightHasMore = true
		}
		if indexCtx.Right[len(indexCtx.Right)-1] < totalPages {
			indexCtx.Last = true
		}
	} else if pageNum == totalPages {
		var startNum int
		if (pageNum - 3) > 0 {
			startNum = pageNum - 3
		}
		indexCtx.Left = pageRange[startNum : pageNum-1]
		if indexCtx.Left[0] > 2 {
			indexCtx.LeftHasMore = true
		}
		if indexCtx.Left[0] > 1 {
			indexCtx.First = true
		}
	} else {
		var startNum int
		if (pageNum - 3) > 0 {
			startNum = pageNum - 3
		}
		indexCtx.Left = pageRange[startNum : pageNum-1]
		indexCtx.Right = pageRange[pageNum:rightEnd]
		if indexCtx.Right[len(indexCtx.Right)-1]+1 < totalPages {
			indexCtx.RightHasMore = true
		}
		if indexCtx.Right[len(indexCtx.Right)-1] < totalPages {
			indexCtx.Last = true
		}
		if indexCtx.Left[0] > 2 {
			indexCtx.LeftHasMore = true
		}
		if indexCtx.Left[0] > 1 {
			indexCtx.First = true
		}
	}
}
func DefaultBaseCtxData() *baseCtxData {
	nowYear, _, _ := time.Now().Date()
	return &baseCtxData{NowYear: nowYear}
}

func DefaultIdxCtxData() *indexContextData {
	return &indexContextData{baseCtxData: *DefaultBaseCtxData(), PageNum: 1, TotalPageNum: 1}
}
func DefaultDetailCtxData() *detailCtxData {
	return &detailCtxData{baseCtxData: *DefaultBaseCtxData()}
}
func DefaultArchivesCtxData() *archivesCtxData {
	return &archivesCtxData{baseCtxData: *DefaultBaseCtxData()}
}
func DefaultTagsCtxData() *tagCtxData {
	return &tagCtxData{baseCtxData: *DefaultBaseCtxData()}
}
func DefaultCategoriesData() *categoryCtxData {
	return &categoryCtxData{baseCtxData: *DefaultBaseCtxData()}
}
func trace(message string) string {
	var pcs [32]uintptr
	n := runtime.Callers(4, pcs[:])

	var str strings.Builder
	str.WriteString(message + "\nTraceback:")
	for _, pc := range pcs[:n] {
		fn := runtime.FuncForPC(pc)
		file, line := fn.FileLine(pc)
		str.WriteString(fmt.Sprintf("\n\t%s:%d", file, line))
	}
	return str.String()
}

type modelFunc func() (map[string]any, error)

func execModel(group *errgroup.Group, dataChan chan map[string]any, f modelFunc) {
	group.Go(func() (err error) {
		defer func() {
			if r := recover(); r != nil {
				message := fmt.Sprintf("errgroup: panic recovered: %s", r)
				err = fmt.Errorf(message)
				log.Error(trace(fmt.Sprintf("errgroup: panic recovered: %s", r)))
			}
		}()

		ret, err := f()
		if err != nil {
			return err
		} else {
			dataChan <- ret
			return nil
		}
	})
}

func makeRange(min, max int) []int {

	a := make([]int, max-min+1)

	for i := range a {
		a[i] = min + i
	}

	return a

}

func baseCount(g *errgroup.Group) (int, int, int, error) {
	dataChan := make(chan map[string]any, 3)
	execModel(g, dataChan,
		func() (map[string]any, error) {
			ret, err := model.GetCategoryCount()
			return map[string]any{"categoryCount": ret}, err
		})
	execModel(g, dataChan,
		func() (map[string]any, error) {
			ret, err := model.GetPostCount()
			return map[string]any{"totalCount": ret}, err
		})
	execModel(g, dataChan,
		func() (map[string]any, error) {
			ret, err := model.GetTagCount()
			return map[string]any{"tagCount": ret}, err
		})
	err := g.Wait()
	close(dataChan)
	if err != nil {
		return 0, 0, 0, err
	}
	dbData := map[string]any{}
	for data := range dataChan {
		for key := range data {
			dbData[key] = data[key]
		}
	}
	tagCount := dbData["tagCount"].(int)
	totalCount := dbData["totalCount"].(int)
	categoryCount := dbData["categoryCount"].(int)
	return tagCount, categoryCount, totalCount, nil
}

func BlogDetail(c *gin.Context) {
	var args detailParam
	var dataChan = make(chan map[string]any, 1)
	if err := c.BindUri(&args); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	postID := args.PostID

	var g errgroup.Group
	execModel(&g, dataChan,
		func() (map[string]any, error) {
			ret, err := model.GetPostDetailWithTagCate(postID)
			return map[string]any{"blogPost": ret}, err
		})
	tagCount, categoryCount, totalCount, err := baseCount(&g)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	err = g.Wait()
	close(dataChan)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	dbData := map[string]any{}
	for data := range dataChan {
		for key := range data {
			dbData[key] = data[key]
		}
	}

	blogPost := dbData["blogPost"].(*model.BlogPost)

	//md := markdown.New(markdown.XHTMLOutput(true))
	//blogPost.Body = md.RenderToString([]byte(blogPost.Body))
	start := time.Now()
	md := goldmark.New(
		// 支持 GFM
		goldmark.WithExtensions(extension.GFM, extension.CJK),
	)
	md.Parser().AddOptions(
		parser.WithAutoHeadingID(),
		parser.WithASTTransformers(
			util.Prioritized(&toc.Transformer{
				Title: "目录",
			}, 100),
		),
	)
	var buf bytes.Buffer
	if err := md.Convert([]byte(blogPost.Body), &buf); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "markdown fail!"})
		return
	}
	blogPost.Views += 1
	err = conf.DB.Model(&model.BlogPost{ID: uint(postID)}).Update("views", gorm.Expr("views+1")).Error
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	blogPost.Body = buf.String()
	log.Info("exec time: ", time.Since(start))
	detailCtx := DefaultDetailCtxData()
	detailCtx.GinCtx = c
	detailCtx.BlogPost = blogPost
	detailCtx.TagCount = tagCount
	detailCtx.PostCount = totalCount
	detailCtx.CategoryCount = categoryCount

	c.HTML(http.StatusOK, "detail.html", detailCtx)
}

func Archives(c *gin.Context) {
	var g errgroup.Group
	dataChan := make(chan map[string]any, 1)
	execModel(&g, dataChan,
		func() (map[string]any, error) {
			ret, err := model.GroupArchive()
			return map[string]any{"archives": ret}, err
		})
	tagCount, categoryCount, totalCount, err := baseCount(&g)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	err = g.Wait()
	close(dataChan)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	dbData := map[string]any{}
	for data := range dataChan {
		for key := range data {
			dbData[key] = data[key]
		}
	}
	archives := dbData["archives"].([]*model.Archive)

	archivesCtx := DefaultArchivesCtxData()
	archivesCtx.GinCtx = c
	archivesCtx.Archives = archives
	archivesCtx.MenuArchive = true
	archivesCtx.TagCount, archivesCtx.CategoryCount, archivesCtx.PostCount = tagCount, categoryCount, totalCount
	c.HTML(http.StatusOK, "archives.html", archivesCtx)
}

func ArchivePosts(c *gin.Context) {
	var args archiveParam
	if err := c.BindUri(&args); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	year := args.Year
	month := args.Month
	var g errgroup.Group
	dataChan := make(chan map[string]any, 1)
	execModel(&g, dataChan,
		func() (map[string]any, error) {
			ret, err := model.ArchivePosts(year, month)
			return map[string]any{"blogPosts": ret}, err
		})
	tagCount, categoryCount, totalCount, err := baseCount(&g)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	err = g.Wait()
	close(dataChan)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	dbData := map[string]any{}
	for data := range dataChan {
		for key := range data {
			dbData[key] = data[key]
		}
	}
	blogPosts := dbData["blogPosts"].([]*model.BlogPost)

	for _, post := range blogPosts {
		md := markdown.New(markdown.XHTMLOutput(true))
		post.Excerpt = md.RenderToString([]byte(post.Excerpt))

	}

	archivePostCtx := DefaultIdxCtxData()
	archivePostCtx.TagCount = tagCount
	archivePostCtx.PostCount = totalCount
	archivePostCtx.BlogPosts = blogPosts
	archivePostCtx.CategoryCount = categoryCount
	archivePostCtx.MenuArchive = true
	archivePostCtx.GinCtx = c
	c.HTML(http.StatusOK, "index.html", archivePostCtx)
}

func Tags(c *gin.Context) {
	var g errgroup.Group
	dataChan := make(chan map[string]any, 1)
	execModel(&g, dataChan,
		func() (map[string]any, error) {
			ret, err := model.TagList()
			return map[string]any{"tags": ret}, err
		})
	tagCount, categoryCount, totalCount, err := baseCount(&g)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	err = g.Wait()
	close(dataChan)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	dbData := map[string]any{}
	for data := range dataChan {
		for key := range data {
			dbData[key] = data[key]
		}
	}
	tags := dbData["tags"].([]*model.TagCount)
	tagsCtx := DefaultTagsCtxData()
	tagsCtx.GinCtx = c
	tagsCtx.Tags = tags
	tagsCtx.MenuTag = true
	tagsCtx.TagCount, tagsCtx.CategoryCount, tagsCtx.PostCount = tagCount, categoryCount, totalCount
	c.HTML(http.StatusOK, "tags.html", tagsCtx)
}

func TagPosts(c *gin.Context) {
	var args tagParam
	if err := c.BindUri(&args); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	tagID := args.TagID
	var g errgroup.Group
	dataChan := make(chan map[string]any, 1)
	execModel(&g, dataChan,
		func() (map[string]any, error) {
			ret, err := model.TagPosts(tagID)
			return map[string]any{"tagPosts": ret}, err
		})
	tagCount, categoryCount, totalCount, err := baseCount(&g)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	err = g.Wait()
	close(dataChan)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	dbData := map[string]any{}
	for data := range dataChan {
		for key := range data {
			dbData[key] = data[key]
		}
	}
	blogPosts := dbData["tagPosts"].([]*model.BlogPost)

	for _, post := range blogPosts {
		md := markdown.New(markdown.XHTMLOutput(true))
		post.Excerpt = md.RenderToString([]byte(post.Excerpt))
	}

	archivePostCtx := DefaultIdxCtxData()
	archivePostCtx.TagCount = tagCount
	archivePostCtx.PostCount = totalCount
	archivePostCtx.BlogPosts = blogPosts
	archivePostCtx.CategoryCount = categoryCount
	archivePostCtx.MenuTag = true
	archivePostCtx.GinCtx = c
	c.HTML(http.StatusOK, "index.html", archivePostCtx)
}

func Categories(c *gin.Context) {
	var g errgroup.Group
	dataChan := make(chan map[string]any, 1)
	execModel(&g, dataChan,
		func() (map[string]any, error) {
			ret, err := model.CategoryList()
			return map[string]any{"categories": ret}, err
		})
	tagCount, categoryCount, totalCount, err := baseCount(&g)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	err = g.Wait()
	close(dataChan)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	dbData := map[string]any{}
	for data := range dataChan {
		for key := range data {
			dbData[key] = data[key]
		}
	}
	categories := dbData["categories"].([]*model.CategoryCount)
	categoryCtx := DefaultCategoriesData()
	categoryCtx.GinCtx = c
	categoryCtx.Categories = categories
	categoryCtx.TagCount, categoryCtx.CategoryCount, categoryCtx.PostCount = tagCount, categoryCount, totalCount
	c.HTML(http.StatusOK, "category.html", categoryCtx)
}

func CategoryPosts(c *gin.Context) {
	var args categoryParam
	if err := c.BindUri(&args); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	categoryID := args.CategoryID
	var g errgroup.Group
	dataChan := make(chan map[string]any, 1)
	execModel(&g, dataChan,
		func() (map[string]any, error) {
			ret, err := model.CategoryPosts(categoryID)
			return map[string]any{"categoryPosts": ret}, err
		})
	tagCount, categoryCount, totalCount, err := baseCount(&g)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	err = g.Wait()
	close(dataChan)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	dbData := map[string]any{}
	for data := range dataChan {
		for key := range data {
			dbData[key] = data[key]
		}
	}
	blogPosts := dbData["categoryPosts"].([]*model.BlogPost)

	for _, post := range blogPosts {
		md := markdown.New(markdown.XHTMLOutput(true))
		post.Excerpt = md.RenderToString([]byte(post.Excerpt))
	}

	categoryPostCtx := DefaultIdxCtxData()
	categoryPostCtx.TagCount = tagCount
	categoryPostCtx.PostCount = totalCount
	categoryPostCtx.BlogPosts = blogPosts
	categoryPostCtx.CategoryCount = categoryCount
	categoryPostCtx.GinCtx = c
	c.HTML(http.StatusOK, "index.html", categoryPostCtx)
}
