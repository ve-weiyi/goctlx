# goctlx - 代码生成工具

sparkinai-cloud 项目的代码生成工具，用于从 API 定义文件和数据库自动生成前端和数据库模型代码，提高开发效率。

## 目录结构

```
goctlx/
├── cmd/                    # 命令行工具实现
│   ├── model/             # Model 代码生成命令
│   │   └── mysql/        # MySQL 模型生成
│   └── web/               # Web 代码生成命令
│       └── ts/           # TypeScript 生成
├── parserx/               # 解析器
│   ├── apispec/          # API 解析器
│   └── dbspec/           # 数据库解析器
├── template/              # 代码模板
│   ├── model/            # Model 模板
│   └── web/              # Web 前端模板
│       └── ts/          # TypeScript 模板
├── testdata/              # 测试数据
├── Makefile               # 构建脚本
├── main.go                # 入口文件
└── go.mod                 # 依赖管理
```

## 功能概述

| 功能         | 说明                                  | 生成内容                       |
|------------|-------------------------------------|----------------------------|
| Model 代码生成 | 从数据库或 SQL 文件生成 GORM 模型              | 数据库模型文件                    |
| Web 代码生成   | 从 `.api` 或 `swagger.json` 生成 TS 代码  | API 接口调用代码和类型定义            |

## 命令结构

```
goctlx
├── model                  # Model 代码生成
│   └── mysql             # MySQL 数据库
│       ├── ddl           # 从 SQL 文件生成
│       └── dsn           # 从数据库连接生成
└── web                    # Web 代码生成
    └── ts                # TypeScript
        ├── api           # 从 .api 文件生成
        └── swagger       # 从 swagger.json 生成
```

## 命令参数说明

| 参数   | 说明       | 示例                                                  |
|------|----------|-----------------------------------------------------|
| `-f` | 输入文件路径   | `/path/to/app.api` 或 `/path/to/swagger.json`        |
| `-t` | 模板目录路径   | `./template/model/model.tpl`                        |
| `-o` | 输出目录路径   | `/path/to/output`                                   |
| `-n` | 输出文件名格式  | `%s.go` / `%v_model.go`                             |
| `-s` | SQL 文件路径 | `/path/to/schema.sql`                               |
| `-u` | 数据库连接字符串 | `root:password@(host:3306)/database`                |

## 快速开始

### 使用 Makefile（推荐）

```bash
# 查看所有可用命令
make help

# 安装依赖
make deps

# 测试命令
make test-model              # 测试 model 生成
make test-web-ts-swagger     # 测试 TypeScript swagger 生成

# Sparkinai 项目命令
make gen-model-sparkinai     # 为 sparkinai 项目生成数据库模型
make gen-web-ts-app          # 为 sparkinai-app 生成 TypeScript API
make gen-web-ts-admin        # 为 sparkinai-admin 生成 TypeScript API

# 清理生成的代码
make clean
```

### 使用命令行

**生成 Model 代码**

```bash
# 从 SQL 文件生成
go run main.go model mysql ddl \
  -s /path/to/schema.sql \
  -t ./template/model/model.tpl \
  -o /path/to/output \
  -n '%v_model.go'

# 从数据库连接生成
go run main.go model mysql dsn \
  -u 'root:password@(127.0.0.1:3306)/database?charset=utf8mb4&parseTime=True&loc=Local' \
  -t ./template/model/model.tpl \
  -o /path/to/output \
  -n '%v_model.go'
```

**生成 TypeScript 代码**

```bash
# 从 .api 文件生成
go run main.go web ts api \
  -f /path/to/app.api \
  -o /path/to/output

# 从 swagger.json 文件生成
go run main.go web ts swagger \
  -f /path/to/swagger.json \
  -o /path/to/output
```

## 工作原理

### 解析器

**ApiSpec Parser** - 解析 `.api` 和 `swagger.json` 文件

- 使用 go-zero 的 API 解析器解析 `.api` 文件
- 使用 go-openapi 解析 `swagger.json` 文件
- 提取接口路径、方法、参数、返回值
- 生成统一的 ApiSpec 结构

**DBSpec Parser** - 解析数据库结构

- 支持从 SQL DDL 文件解析
- 支持从数据库 DSN 连接解析
- 提取表结构、字段、索引等信息

### 代码生成流程

1. 解析输入文件（`.api`、`swagger.json` 或数据库）
2. 转换为统一的数据结构
3. 根据模板生成目标代码
4. 输出到指定目录

## 技术特点

- ✅ 基于 Cobra 构建命令行工具
- ✅ 支持自定义模板，灵活扩展
- ✅ 支持多种代码生成场景
- ✅ 自动化生成，减少重复代码编写
- ✅ 统一代码风格，提高代码质量

## 使用场景

1. **数据库变更** - 修改 SQL 文件后，运行 `make gen-model-sparkinai` 更新模型
2. **前后端协作** - 运行 `make gen-web-ts-app` 或 `make gen-web-ts-admin` 生成前端 TypeScript 代码
3. **已有 Swagger** - 从现有 Swagger 文档生成 TypeScript 代码

## 注意事项

1. 生成代码前建议先备份或提交现有代码
2. 生成的代码可能需要手动调整部分逻辑
3. 模板文件位于 `template/` 目录，可根据需求自定义
4. 生成的文件会覆盖同名文件，请谨慎操作
