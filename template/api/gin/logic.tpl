package {{.Package}}

import (
	{{- range .Imports}}
	{{.}}
	{{- end}}
)

type {{.Group}}Logic struct {
	svcCtx *svctx.ServiceContext
}

func New{{.Group}}Logic(svcCtx *svctx.ServiceContext) *{{.Group}}Logic {
	return &{{.Group}}Logic{
		svcCtx: svcCtx,
	}
}

{{- range .GroupRoutes}}
{{- range .Routes}}

// {{.Doc}}
func (l *{{$.Group}}Logic) {{.Handler}}(reqCtx *request.Context
	{{- if .Request }}, req {{pkgTypes .Request}}{{ end }}) (
	{{- if .Response }}resp {{pkgTypes .Response}}, {{ end -}}
	err error) {
	// TODO: implement business logic

	return
}
{{- end}}
{{- end}}
