# gm-cli
golang通过连接数据库生成表对应的model

## 使用
创建配置文件config.ini
```go
[mysql]
link = "your_username:your_password@tcp(your_hostname:3306)/your_database"
prefix = "go_"
removePrefix = "t_,b_,xh_"
tables = "lxt_table1,lxt_table2"
```
执行命令
```go
go run main.go
```
生成文件的文件在generated中