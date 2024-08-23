.PHONY: docker
docker:
	@rm webook || true
	@go mod tidy
	@GOOS=linux GOARCH=arm go build -o webook . && chmod +x webook # 添加执行权限
	@docker rmi -f ting17/webook:v0.0.2
	@docker build -t ting17/webook:v0.0.2 .

# make docker

# chmod 的全称是 change mode。
# 这是一个 Unix 和 Linux 系统上的命令，用于改变文件或目录的权限模式。
# 通过 chmod 命令，你可以设置哪些用户可以读取、写入或执行特定的文件或目录。

# +x 是 chmod 命令中的一个选项，用于为文件或目录添加执行权限（execute permission）。
# +: 表示添加权限。
# x: 表示执行权限（execute permission）。