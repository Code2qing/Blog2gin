package controllers

import (
	"Blog2Gin/conf"
	"Blog2Gin/model"
	"Blog2Gin/server/common"
	"Blog2Gin/server/forms"
	"fmt"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
	"net/http"
	"time"
)

func BlogIndex(c *gin.Context) {
	var args forms.IndexParam
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
	//go common.ExecModel(blogPostChan, errCH, func() (any, error) { return model.GetPostsByPage(pageNum, conf.IndexPageSize) })
	//go common.ExecModel(blogCtPostChan, errCH, func() (any, error) { return model.GetPostsByCategory(12) })
	//go common.ExecModel(totalCountChan, errCH, func() (any, error) { return model.GetPostCount() })
	common.ExecModel(&g, dataChan,
		func() (map[string]any, error) {
			ret, err := model.GetPostsByPage(pageNum, conf.IndexPageSize)
			return map[string]any{"blogPosts": ret}, err
		})
	//common.ExecModel(&g, dataChan,
	//	func() (map[string]any, error) {
	//		ret, err := model.GetPostsByCategory(1)
	//		return map[string]any{"blog_category_posts": ret}, err
	//	})
	tagCount, categoryCount, totalCount, err := common.BaseCount(&g)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	//g.Go(common.ExecModel(
	//	dataChan,
	//	func() (map[string]any, error) {
	//		ret, err := model.GetPostsByPage(pageNum, conf.IndexPageSize)
	//		return map[string]any{"blog_posts": ret}, err
	//	}))
	//
	//g.Go(common.ExecModel(
	//	dataChan,
	//	func() (map[string]any, error) {
	//		ret, err := model.GetPostsByCategory(1)
	//		return map[string]any{"blog_category_posts": ret}, err
	//	}))
	//g.Go(common.ExecModel(
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
	//for _, post := range blogPosts {
	//	//markdown := goldmark.New(
	//	//	// 支持 GFM
	//	//	goldmark.WithExtensions(extension.GFM),
	//	//)
	//	//var buf bytes.Buffer
	//	//if err := markdown.Convert([]byte(post.Excerpt), &buf); err != nil {
	//	//	c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "markdown fail!"})
	//	//}
	//	//post.Excerpt = buf.String()
	//
	//	//post.Excerpt = string(github_flavored_markdown.Markdown([]byte(post.Excerpt)))
	//
	//	//post.Excerpt = string(blackfriday.Run([]byte(post.Excerpt), blackfriday.WithExtensions()))
	//
	//	md := markdown.New(markdown.XHTMLOutput(true))
	//	post.Excerpt = md.RenderToString([]byte(post.Excerpt))
	//
	//}
	log.Infof("exec time is: %s", time.Since(start))
	indexCtx := common.DefaultIdxCtxData()
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
		common.PaginationData(indexCtx, pageNum, totalPages)
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

//func common.ExecModel(ch chan any, errCH chan error, f func() (any, error)) {
//	ret, err := f()
//	if err != nil {
//		errCH <- err
//	} else {
//		ch <- ret
//	}
//}

func BlogDetail(c *gin.Context) {
	var args forms.DetailParam
	var dataChan = make(chan map[string]any, 1)
	if err := c.BindUri(&args); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	postID := args.PostID

	var g errgroup.Group
	common.ExecModel(&g, dataChan,
		func() (map[string]any, error) {
			ret, err := model.GetPostDetailWithTagCate(postID)
			return map[string]any{"blogPost": ret}, err
		})
	tagCount, categoryCount, totalCount, err := common.BaseCount(&g)
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
	//start := time.Now()
	//md := goldmark.New(
	//	// 支持 GFM
	//	goldmark.WithExtensions(extension.GFM, extension.CJK),
	//)
	//md.Parser().AddOptions(
	//	parser.WithAutoHeadingID(),
	//	parser.WithASTTransformers(
	//		util.Prioritized(&toc.Transformer{
	//			Title: "目录",
	//		}, 100),
	//	),
	//)
	//var buf bytes.Buffer
	//if err := md.Convert([]byte(blogPost.Body), &buf); err != nil {
	//	c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "markdown fail!"})
	//	return
	//}
	//blogPost.Body = buf.String()
	//log.Info("exec time: ", time.Since(start))
	blogPost.Views += 1
	err = conf.DB.Model(&model.BlogPost{ID: uint(postID)}).Update("views", gorm.Expr("views+1")).Error
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	detailCtx := common.DefaultDetailCtxData()
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
	common.ExecModel(&g, dataChan,
		func() (map[string]any, error) {
			ret, err := model.GroupArchive()
			return map[string]any{"archives": ret}, err
		})
	tagCount, categoryCount, totalCount, err := common.BaseCount(&g)
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

	archivesCtx := common.DefaultArchivesCtxData()
	archivesCtx.GinCtx = c
	archivesCtx.Archives = archives
	archivesCtx.MenuArchive = true
	archivesCtx.TagCount, archivesCtx.CategoryCount, archivesCtx.PostCount = tagCount, categoryCount, totalCount
	c.HTML(http.StatusOK, "archives.html", archivesCtx)
}

func ArchivePosts(c *gin.Context) {
	var args forms.ArchiveParam
	if err := c.BindUri(&args); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	year := args.Year
	month := args.Month
	var g errgroup.Group
	dataChan := make(chan map[string]any, 1)
	common.ExecModel(&g, dataChan,
		func() (map[string]any, error) {
			ret, err := model.ArchivePosts(year, month)
			return map[string]any{"blogPosts": ret}, err
		})
	tagCount, categoryCount, totalCount, err := common.BaseCount(&g)
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

	//for _, post := range blogPosts {
	//	md := markdown.New(markdown.XHTMLOutput(true))
	//	post.Excerpt = md.RenderToString([]byte(post.Excerpt))
	//
	//}

	archivePostCtx := common.DefaultIdxCtxData()
	archivePostCtx.TagCount = tagCount
	archivePostCtx.PostCount = totalCount
	archivePostCtx.BlogPosts = blogPosts
	archivePostCtx.CategoryCount = categoryCount
	archivePostCtx.MenuArchive = true
	archivePostCtx.GinCtx = c
	archivePostCtx.Title = fmt.Sprintf("归档 - %d年%d月", year, month)
	c.HTML(http.StatusOK, "index.html", archivePostCtx)
}

func Tags(c *gin.Context) {
	var g errgroup.Group
	dataChan := make(chan map[string]any, 1)
	common.ExecModel(&g, dataChan,
		func() (map[string]any, error) {
			ret, err := model.TagList()
			return map[string]any{"tags": ret}, err
		})
	tagCount, categoryCount, totalCount, err := common.BaseCount(&g)
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
	tagsCtx := common.DefaultTagsCtxData()
	tagsCtx.GinCtx = c
	tagsCtx.Tags = tags
	tagsCtx.MenuTag = true
	tagsCtx.TagCount, tagsCtx.CategoryCount, tagsCtx.PostCount = tagCount, categoryCount, totalCount
	c.HTML(http.StatusOK, "tags.html", tagsCtx)
}

func TagPosts(c *gin.Context) {
	var args forms.TagParam
	if err := c.BindUri(&args); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	tagID := args.TagID
	var g errgroup.Group
	dataChan := make(chan map[string]any, 2)
	common.ExecModel(&g, dataChan,
		func() (map[string]any, error) {
			ret, err := model.TagPosts(tagID)
			return map[string]any{"tagPosts": ret}, err
		})
	common.ExecModel(&g, dataChan,
		func() (map[string]any, error) {
			ret, err := model.TagInfo(tagID)
			return map[string]any{"tag": ret}, err
		})
	tagCount, categoryCount, totalCount, err := common.BaseCount(&g)
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
	tagInfo := dbData["tag"].(*model.Tag)

	//for _, post := range blogPosts {
	//	md := markdown.New(markdown.XHTMLOutput(true))
	//	post.Excerpt = md.RenderToString([]byte(post.Excerpt))
	//}

	tagPostCtx := common.DefaultIdxCtxData()
	tagPostCtx.TagCount = tagCount
	tagPostCtx.PostCount = totalCount
	tagPostCtx.BlogPosts = blogPosts
	tagPostCtx.CategoryCount = categoryCount
	tagPostCtx.MenuTag = true
	tagPostCtx.GinCtx = c
	tagPostCtx.Title = "标签 - " + tagInfo.TagName
	c.HTML(http.StatusOK, "index.html", tagPostCtx)
}

func Categories(c *gin.Context) {
	var g errgroup.Group
	dataChan := make(chan map[string]any, 1)
	common.ExecModel(&g, dataChan,
		func() (map[string]any, error) {
			ret, err := model.CategoryList()
			return map[string]any{"categories": ret}, err
		})
	tagCount, categoryCount, totalCount, err := common.BaseCount(&g)
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
	categoryCtx := common.DefaultCategoriesData()
	categoryCtx.GinCtx = c
	categoryCtx.Categories = categories
	categoryCtx.TagCount, categoryCtx.CategoryCount, categoryCtx.PostCount = tagCount, categoryCount, totalCount
	c.HTML(http.StatusOK, "category.html", categoryCtx)
}

func CategoryPosts(c *gin.Context) {
	var args forms.CategoryParam
	if err := c.BindUri(&args); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	categoryID := args.CategoryID
	var g errgroup.Group
	dataChan := make(chan map[string]any, 1)
	common.ExecModel(&g, dataChan,
		func() (map[string]any, error) {
			ret, err := model.CategoryPosts(categoryID)
			return map[string]any{"categoryPosts": ret}, err
		})
	tagCount, categoryCount, totalCount, err := common.BaseCount(&g)
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

	//for _, post := range blogPosts {
	//	md := markdown.New(markdown.XHTMLOutput(true))
	//	post.Excerpt = md.RenderToString([]byte(post.Excerpt))
	//}

	categoryPostCtx := common.DefaultIdxCtxData()
	categoryPostCtx.TagCount = tagCount
	categoryPostCtx.PostCount = totalCount
	categoryPostCtx.BlogPosts = blogPosts
	categoryPostCtx.CategoryCount = categoryCount
	categoryPostCtx.GinCtx = c
	c.HTML(http.StatusOK, "index.html", categoryPostCtx)
}
