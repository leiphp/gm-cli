# gm-cli
golang通过连接数据库生成表对应的model

## 使用
在项目根目录下创建配置文件config.ini
```go
[mysql]
link = "your_username:your_password@tcp(your_hostname:3306)/your_database"
prefix = "go_"
removePrefix = "t_,b_,xh_"
tables = "lxt_table1,lxt_table2"
```

在Release下载gm_windows_amd64.exe，重命名gm.exe。配置环境变量执行gm即可，注意要在config.ini同级目录执行
```go
gm
```
生成文件的文件在model中，也可以在config.ini配置