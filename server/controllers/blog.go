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
	var dataChan = make(chan map[string]any, 4)

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
	execModel(&g, dataChan,
		func() (map[string]any, error) {
			ret, err := model.GetCategoryCount()
			return map[string]any{"categoryCount": ret}, err
		})
	execModel(&g, dataChan,
		func() (map[string]any, error) {
			ret, err := model.GetPostCount()
			return map[string]any{"totalCount": ret}, err
		})
	execModel(&g, dataChan,
		func() (map[string]any, error) {
			ret, err := model.GetTagCount()
			return map[string]any{"tagCount": ret}, err
		})
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

	err := g.Wait()
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
	totalCountV, ok := dbData["totalCount"]
	blogPostsV, ok := dbData["blogPosts"]
	categoryCountV, ok := dbData["categoryCount"]
	tagCountV, ok := dbData["tagCount"]
	if !ok {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "not total_count!"})
	}
	tagCount := tagCountV.(int)
	totalCount := totalCountV.(int)
	categoryCount := categoryCountV.(int)
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

func DefaultIdxCtxData() *indexContextData {
	nowYear, _, _ := time.Now().Date()
	return &indexContextData{baseCtxData: baseCtxData{NowYear: nowYear}, PageNum: 1, TotalPageNum: 1}
}
func DefaultDetailCtxData() *detailCtxData {
	nowYear, _, _ := time.Now().Date()
	return &detailCtxData{baseCtxData: baseCtxData{NowYear: nowYear}}
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

func BlogDetail(c *gin.Context) {
	var args detailParam
	if err := c.BindUri(&args); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	postID := args.PostID
	blogPost, err := model.GetPostDetailWithTagCate(postID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
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
	}
	blogPost.Body = buf.String()
	log.Info("exec time: ", time.Since(start))
	detailCtx := DefaultDetailCtxData()
	detailCtx.GinCtx = c
	detailCtx.BlogPost = blogPost
	c.HTML(http.StatusOK, "detail.html", detailCtx)
}
