package templates

import (
	"embed"
	"github.com/gin-contrib/multitemplate"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/render"
	"html/template"
	"io/fs"
	"path/filepath"
)

//go:embed *
var tmplFS embed.FS

type CustomRenderer interface {
	render.HTMLRender
	multitemplate.Renderer
	AddFromFsFilesFuncs(name string, funcMap template.FuncMap, fs fs.FS, files ...string) *template.Template
}

type customRender struct {
	multitemplate.Render
}

func (r customRender) AddFromFsFilesFuncs(name string, funcMap template.FuncMap, fs fs.FS, files ...string) *template.Template {
	tname := filepath.Base(files[0])
	tmpl := template.Must(template.New(tname).Funcs(funcMap).ParseFS(fs, files...))
	r.Add(name, tmpl)
	return tmpl
}

func NewCustomRender() CustomRenderer {
	return customRender{Render: make(multitemplate.Render)}
}

func loadTemplates() CustomRenderer {
	//r := multitemplate.NewRenderer()
	r := NewCustomRender()
	{
		blogLayouts, err := fs.Glob(tmplFS, "blog/base/*.html")
		if err != nil {
			panic(err.Error())
		}

		blogIncludes, err := fs.Glob(tmplFS, "blog/*.html")
		if err != nil {
			panic(err.Error())
		}

		// Generate our templates map from our layouts/ and includes/ directories
		for _, include := range blogIncludes {
			layoutCopy := make([]string, len(blogLayouts))
			copy(layoutCopy, blogLayouts)
			files := append(layoutCopy, include)
			r.AddFromFsFilesFuncs(filepath.Base(include), template.FuncMap{
				"unescaped": unescaped,
				"incr":      incr,
			}, tmplFS, files...)
		}
	}

	{
		editorLayouts, err := fs.Glob(tmplFS, "editor/base/*.html")
		if err != nil {
			panic(err.Error())
		}
		editorIncludes, err := fs.Glob(tmplFS, "editor/*.html")
		if err != nil {
			panic(err.Error())
		}
		for _, include := range editorIncludes {
			layoutCopy := make([]string, len(editorLayouts))
			copy(layoutCopy, editorLayouts)
			files := append(layoutCopy, include)
			r.AddFromFsFilesFuncs(filepath.Base(include), template.FuncMap{
				"unescaped": unescaped,
			}, tmplFS, files...)
		}
	}

	return r
}

func unescaped(x string) interface{} { return template.HTML(x) }

func incr(i int) int { return i + 1 }

func InitTemplate(router *gin.Engine) {
	router.HTMLRender = loadTemplates()
}
