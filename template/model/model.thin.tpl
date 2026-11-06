package model

import (
	"context"

	"gorm.io/gorm"
)

var _ {{.UpperStartCamelName}}Model = (*default{{.UpperStartCamelName}}Model)(nil)

type (
	// 接口定义
	{{.UpperStartCamelName}}Model interface {
		TableName() string
		// 在事务中操作
		WithTx(tx *gorm.DB) (out {{.UpperStartCamelName}}Model)
		// 插入
		Insert(ctx context.Context, in *{{.UpperStartCamelName}}) (rows int64, err error)
		// 删除
		Delete(ctx context.Context, id int64) (rows int64, err error)
		// 整行更新（含零值，按主键）
		Update(ctx context.Context, in *{{.UpperStartCamelName}}) (rows int64, err error)
		// 插入或整行更新（按主键）
		Upsert(ctx context.Context, in *{{.UpperStartCamelName}}) (rows int64, err error)
		// 查询
		FindById(ctx context.Context, id int64) (out *{{.UpperStartCamelName}}, err error)
        {{- if .UniqueFields}}
		// add extra method in here
        {{- end}}
        {{- range $key, $value := .UniqueFields}}
            FindOneBy{{ funcFieldsKey $value}}(ctx context.Context, {{funcFieldsKeyVar $value}}) (out *{{$.UpperStartCamelName}},err error)
        {{- end}}
	}

	// 表字段定义
	{{.UpperStartCamelName}} struct {
	{{range .Fields}}
        {{.Name}} {{.Type}} `{{.Tag}}` {{if .Comment}}// {{.Comment}}{{end}}
	{{- end}}
	}

    // 接口实现
    default{{.UpperStartCamelName}}Model struct {
        baseModel[{{.UpperStartCamelName}}]
    }
)

func New{{.UpperStartCamelName}}Model(db *gorm.DB) {{.UpperStartCamelName}}Model {
	return &default{{.UpperStartCamelName}}Model{newBaseModel[{{.UpperStartCamelName}}](db, "`{{.SnakeName}}`")}
}

// 在事务中操作
func (m *default{{.UpperStartCamelName}}Model) WithTx(tx *gorm.DB) (out {{.UpperStartCamelName}}Model) {
	return &default{{.UpperStartCamelName}}Model{newBaseModel[{{.UpperStartCamelName}}](tx, m.table)}
}

{{- if .UniqueFields}}
// add extra method in here
{{- end}}
{{- range $key, $value := .UniqueFields}}
func (m *default{{$.UpperStartCamelName}}Model) FindOneBy{{ funcFieldsKey $value}}(ctx context.Context, {{funcFieldsKeyVar $value}}) (out *{{$.UpperStartCamelName}},err error) {
	db := m.scope(ctx)

	err = db.Where("{{funcFieldsKeyCond $value}}", {{funcFieldsKeyCondVar $value}}).First(&out).Error
	if err != nil {
		return nil, err
	}

	return out, nil
}
{{ end }}
