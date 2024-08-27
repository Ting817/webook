# 基础镜像
FROM ubuntu:latest

# 把编译后的打包进来这个镜像，放到工作目录里 /app
RUN mkdir -p /data
COPY webook /data/webook
WORKDIR /data

# CMD 是执行命令
# ENTRYPOINT 最佳
CMD ["sh", "-c", "/data/webook"]