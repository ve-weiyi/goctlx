package {{ .Package }}

import (
	"github.com/gin-gonic/gin"

	{{- range .Imports}}
	{{.}}
	{{- end}}
)

func RegisterHandlers(r *gin.RouterGroup, svcCtx *svctx.ServiceContext) {
	{{- range .Groups}}
	router.New{{.}}Router(svcCtx).Register(r)
	{{- end}}
}
