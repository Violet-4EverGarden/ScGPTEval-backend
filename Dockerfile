FROM golang:1.20.5-alpine3.18 AS builder

# 设置环境变量
ENV GO111MODULE=on \
GOPROXY=https://goproxy.cn,direct \
CGO_ENABLED=0 \
GOOS=linux \
GOARCH=amd64

# 移动到工作目录：/build
WORKDIR /build

# 添加执行程序的用户
RUN adduser -u 10001 -D app-runner

# 复制项目中的 go.mod 和 go.sum文件并下载依赖信息
COPY go.mod .
COPY go.sum .
RUN go mod download

# 将代码复制到容器中
COPY . .

# 将我们的代码编译成二进制可执行文件 bubble
RUN go build -a -o scgpt_eval .

# 接下来创建一个最终镜像
FROM ubuntu:20.04 AS final

# 设置容器内时区
RUN ln -fs /usr/share/zoneinfo/Asia/Shanghai /etc/localtime

# 更新包列表
RUN apt-get update \
		&& apt-get install tzdata -y \
		&& apt-get clean

# 设置工作目录，后续默认启动该目录下的程序
WORKDIR /app

# 从builder镜像的工作目录中把可执行文件、配置文件等内容拷贝到当前目录
COPY --from=builder /build /app

# 一些密钥和加密证书等(可能存在吧)
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# 声明容器内暴露端口为8089
EXPOSE 8089

# 用app-runner启动而不是root
USER app-runner

# 设置容器启动时执行的命令
ENTRYPOINT ["./scgpt_eval", "-c", "conf/config.yaml"]