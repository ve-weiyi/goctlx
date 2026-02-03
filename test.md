生成gin代码（测试）

```sh
go run main.go api gin api \
    -f ./testdata/test.api \
    -t ./template/api/gin \
    -o ./runtime/api \
    -c ./svctx \
    -n '%s.go'
    
```

生成gin代码（blog、admin）

```sh
go run main.go api gin api \
    -f ../blog-gozero/service/api/blog/proto/blog.api \
    -t ./template/api/gin \
    -o ../blog-gin/api/blog \
    -c github.com/ve-weiyi/ve-blog-golang/blog-gin/svctx \
    -n '%s.go'
    
    
go run main.go api gin api \
    -f ../blog-gozero/service/api/admin/proto/admin.api \
    -t ./template/api/gin \
    -o ../blog-gin/api/admin \
    -c github.com/ve-weiyi/ve-blog-golang/blog-gin/svctx \
    -n '%s.go'
```

ddl生成model

```sh
go run main.go model mysql ddl \
    -s ../blog-veweiyi-init.sql \
    -t ./template/model/model.tpl \
    -o ../blog-gozero/service/model \
    -n '%v_model.go'
```

测试model
```sh
go run main.go model mysql ddl \
    -s /Users/weiyi/Github/sparkinai/sparkinai-cloud/sparkinai-log.sql \
    -t ./template/model/model.tpl \
    -o /Users/weiyi/Github/sparkinai/sparkinai-cloud/service/model \
    -n '%v_model.go'
```

生成admin的api文件


```sh
go run main.go web ts api \
  -f /Users/weiyi/Github/sparkinai/sparkinai-cloud/service/api/admin/proto/admin.api \
  -t ./template/web/ts \
  -o /Users/weiyi/Github/sparkinai/sparkinai-admin/src/api \
  -n '%v.ts'
```


```sh
go run main.go web ts api \
  -f /Users/weiyi/Github/sparkinai/sparkinai-cloud/service/api/app/proto/app.api \
  -t ./template/web/ts \
  -o /Users/weiyi/Github/sparkinai/sparkinai-app/src/api \
  -n '%v.ts'
```



```sh
go run main.go web ts api \
  -f ./testdata/test.api \
  -t ./template/web/ts \
  -o ./runtime/testdata/api \
  -n '%v.ts'
```

```sh
go run main.go web ts api \
  -f ../blog-gozero/service/api/blog/proto/blog.api \
  -t ./template/web/ts \
  -o ../../ve-blog-naive/src/api \
  -n '%v.ts'
```
