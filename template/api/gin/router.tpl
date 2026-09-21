package {{.Package}}

import (
	"github.com/gin-gonic/gin"

	{{- range .Imports}}
	{{.}}
	{{- end}}
)

type {{.Group}}Router struct {
	svcCtx *svctx.ServiceContext
}

func New{{.Group}}Router(svcCtx *svctx.ServiceContext) *{{.Group}}Router {
	return &{{.Group}}Router{
		svcCtx: svcCtx,
	}
}

func (r *{{.Group}}Router) Register(rg *gin.RouterGroup) {
	{{- range .GroupRoutes}}
	{
		g := rg.Group("{{.Prefix}}")
		{{- range .Middleware}}
		g.Use(r.svcCtx.{{.}})
		{{- end }}

		h := handler.New{{$.Group}}Handler(r.svcCtx)
		{{- range .Routes}}
		// {{.Doc}}
		g.{{.Method}}("{{.Path}}", h.{{.Handler}})
		{{- end}}
	}
	{{- end}}
}
