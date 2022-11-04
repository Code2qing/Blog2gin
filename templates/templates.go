package templates

import (
	"github.com/gin-contrib/multitemplate"
	"github.com/gin-gonic/gin"
	"html/template"
	"path/filepath"
)

func loadTemplates(templatesDir string) multitemplate.Renderer {
	r := multitemplate.NewRenderer()

	{
		blogLayouts, err := filepath.Glob(templatesDir + "/blog/base/*.html")
		if err != nil {
			panic(err.Error())
		}

		blogIncludes, err := filepath.Glob(templatesDir + "/blog/*.html")
		if err != nil {
			panic(err.Error())
		}

		// Generate our templates map from our layouts/ and includes/ directories
		for _, include := range blogIncludes {
			layoutCopy := make([]string, len(blogLayouts))
			copy(layoutCopy, blogLayouts)
			files := append(layoutCopy, include)
			r.AddFromFilesFuncs(filepath.Base(include), template.FuncMap{
				"unescaped": unescaped,
				"incr":      incr,
			}, files...)
		}
	}

	{
		editorLayouts, err := filepath.Glob(templatesDir + "/editor/base/*.html")
		if err != nil {
			panic(err.Error())
		}
		editorIncludes, err := filepath.Glob(templatesDir + "/editor/*.html")
		if err != nil {
			panic(err.Error())
		}
		for _, include := range editorIncludes {
			layoutCopy := make([]string, len(editorLayouts))
			copy(layoutCopy, editorLayouts)
			files := append(layoutCopy, include)
			r.AddFromFilesFuncs(filepath.Base(include), template.FuncMap{
				"unescaped": unescaped,
			}, files...)
		}
	}

	return r
}

func unescaped(x string) interface{} { return template.HTML(x) }

func incr(i int) int { return i + 1 }

func InitTemplate(router *gin.Engine) {
	router.HTMLRender = loadTemplates("templates")
}
