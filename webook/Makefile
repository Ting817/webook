.PHONY: docker
docker:
	@rm webook || true
	@GOOS=linux GOARCH=arm go build -o webook .
	@docker rm -f ting17/webook:v0.0.1
	@docker build -t ting17/webook:v0.0.1 .

# make docker