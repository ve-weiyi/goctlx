.PHONY: help deps test-model test-web-ts-swagger gen-model-sparkinai gen-web-ts-app gen-web-ts-admin clean

help:
	@echo "可用命令："
	@echo "  deps                  - 安装依赖"
	@echo "  test-model            - 测试 model 生成"
	@echo "  test-web-ts-swagger   - 测试 TypeScript swagger 生成"
	@echo "  gen-model-sparkinai   - 为 sparkinai 项目生成数据库模型"
	@echo "  gen-web-ts-app        - 为 sparkinai-app 生成 TypeScript API"
	@echo "  gen-web-ts-admin      - 为 sparkinai-admin 生成 TypeScript API"
	@echo "  clean                 - 清理生成的代码"

# 安装依赖
deps:
	go mod tidy

# 测试 model 生成
test-model:
	go run main.go model mysql ddl \
		-s ./testdata/t_user.sql \
		-t ./template/model/model.tpl \
		-o ./runtime/model \
		-n '%v_model.go'

# 测试 TypeScript swagger 生成
test-web-ts-swagger:
	go run main.go web ts swagger \
		-f ./testdata/test.json \
		-o ./runtime/web/ts

# 测试 TypeScript api 生成
test-web-ts-api:
	go run main.go web ts api \
		-f /Users/weiyi/Github/veweiyi/sparkinai/sparkinai-cloud/service/admin/api/proto/admin.api \
		-o ./runtime/web/ts

# 为 sparkinai 项目生成数据库模型
gen-model-sparkinai:
	go run main.go model mysql ddl \
		-s /Users/weiyi/Github/veweiyi/sparkinai/sparkinai-cloud/sparkinai.sql \
		-t ./template/model/model.tpl \
		-o /Users/weiyi/Github/veweiyi/sparkinai/sparkinai-cloud/service/app/model \
		-n '%v_model.go'

# 为 sparkinai-app 生成 TypeScript API
gen-web-ts-app:
	go run main.go web ts api \
		-f /Users/weiyi/Github/veweiyi/sparkinai/sparkinai-cloud/service/app/api/proto/app.api \
		-o /Users/weiyi/Github/veweiyi/sparkinai/sparkinai-app/src/api

# 为 sparkinai-admin 生成 TypeScript API
gen-web-ts-admin:
	go run main.go web ts api \
		-f /Users/weiyi/Github/veweiyi/sparkinai/sparkinai-cloud/service/admin/api/proto/admin.api \
		-o /Users/weiyi/Github/veweiyi/sparkinai/sparkinai-admin/src/api

# 清理生成的代码
clean:
	rm -rf ./runtime
