package {{.Package}}

import (
	{{- range .Imports}}
	{{.}}
	{{- end}}
)

type {{.Group}}Handler struct {
	svcCtx *svctx.ServiceContext
}

func New{{.Group}}Handler(svcCtx *svctx.ServiceContext) *{{.Group}}Handler {
	return &{{.Group}}Handler{
		svcCtx: svcCtx,
	}
}

{{- range .GroupRoutes}}
{{- $prefix := .Prefix}}
{{- range .Routes}}

// {{.Doc}}
{{- if .Request}}
// @Param data body {{commentTypes .Request}} true "请求参数"
{{- end}}
{{- if .Response}}
// @Success 200 {object} {{commentTypes .Response}} "返回信息"
{{- end}}
// @Router {{$prefix}}{{.Path}} [{{.Method}}]
func (h *{{$.Group}}Handler) {{.Handler}}(c *gin.Context) {
	reqCtx, err := request.ParseRequestContext(c)
	if err != nil {
		response.ResponseError(c, err)
		return
	}

	{{- if .Request }}
	var req {{pkgTypes .Request}}
	err = request.ShouldBind(c, &req)
	if err != nil {
		response.ResponseError(c, err)
		return
	}
	{{- end }}

	l := logic.New{{$.Group}}Logic(h.svcCtx)
	{{- if .Response }}
	resp, err := l.{{.Handler}}(reqCtx{{if .Request}}, req{{end}})
	{{- else }}
	err = l.{{.Handler}}(reqCtx{{if .Request}}, req{{end}})
	{{- end }}
	if err != nil {
		response.ResponseError(c, err)
		return
	}

	{{- if .Response }}
	response.ResponseOk(c, resp)
	{{- else }}
	response.ResponseOk(c, nil)
	{{- end }}
}
{{- end}}
{{- end}}
