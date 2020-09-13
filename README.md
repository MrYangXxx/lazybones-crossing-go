# lazybones-crossing-go
# 项目介绍
本项目使用了 Hiboot云原生应用框架

>Hiboot 是一款用Go语言实现的高性能网络及命令行应用框架。Hiboot提供了Web MVC框架，支持控制反转（依赖注入），拦截器，依赖包自动配置（类似Spring Boot的Starter）。

更多信息请查看官方文档：https://hiboot.hidevops.io/cn/

# 开发指引

详情请看官方文档: https://hiboot.hidevops.io/cn/web-app/

0.安装依赖
> 根目录下执行 go mod tidy
> 可能需要开启go module，环境变量添加 GO111MODULE=on
>
1.编译
>go build -ldflags -w //ldflags指定编译参数，-w为去掉调试信息
>
>跨平台: 
>- set GOARCH=amd64    //cpu架构，一般为amd64,386,arm
>- set GOOS=linux      //操作系统，一般为windows,linux,darwin
>- go build -ldflags -w 
>
>压缩
>- 对应平台安装[upx](https://github.com/upx/upx/releases)
>- 执行 `upx -9 -k projectName`