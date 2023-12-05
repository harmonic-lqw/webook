@echo off

REM 删除 webook 文件，忽略文件不存在的错误
del webook 2>nul

REM 强制删除 Docker 镜像
docker rmi -f harmonic/webook:v0.0.1 2>nul

REM 运行 go mod tidy
go mod tidy

REM 使用交叉编译，在 ARM 架构的 Linux 操作系统上构建可执行文件 webook
set GOOS=linux
set GOARCH=arm
go build -tags=k8s -o webook .

REM 构建 Docker 镜像
docker build -t harmonic/webook:v0.0.1 .