# 你可以直接执行 make 命令，也可以单独的命令复制到控制台。
# 注意，如果你是 Windows 并且不是在 WSL 下，
# 要注意文件分隔符使用 Windows 的分隔符。
.PHONY: generate
generate:
	@make mock

.PHONY: mock
mock:
	@go generate -tags=wireinject ./...
	@go mod tidy

.PHONY: grpc
grpc:
	# 这句命令要跑到 webook 上一层目录中，编译 buf.gen.yaml
	@buf generate webook/api/proto

.PHONY: e2e
e2e:
	@docker compose -f webook/docker-compose.yaml down
	@docker compose -f webook/docker-compose.yaml up -d
	@go test -race ./webook/... -tags=e2e
	@docker compose -f webook/docker-compose.yaml down

.PHONY: e2e_up
e2e_up:
	@docker compose -f webook/docker-compose.yaml up -d

.PHONY: e2e_down
e2e_down:
	@docker compose -f webook/docker-compose.yaml down