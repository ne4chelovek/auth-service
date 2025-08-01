include .env
export $(shell sed 's/=.*//' .env)

# Пути
LOCAL_BIN := $(CURDIR)/bin
LOCAL_MIGRATION_DIR := ./migrations  # Добавлено, если не определено в .env
LOCAL_MIGRATION_DSN := host=localhost port=${PG_PORT} dbname=${PG_DATABASE_NAME} user=${PG_USER} password=${PG_PASSWORD} sslmode=disable

install-deps:
	GOBIN=$(LOCAL_BIN) go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28.1
	GOBIN=$(LOCAL_BIN) go install -mod=mod google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2
	GOBIN=$(LOCAL_BIN) go install github.com/pressly/goose/v3/cmd/goose@v3.15.1
	GOBIN=$(LOCAL_BIN) go install github.com/envoyproxy/protoc-gen-validate@v1.2.1
	GOBIN=$(LOCAL_BIN) go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@v2.26.2
	GOBIN=$(LOCAL_BIN) go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@v2.20.0
	GOBIN=$(LOCAL_BIN) go install github.com/rakyll/statik@v0.1.7

get-deps:
	go get -u google.golang.org/protobuf/cmd/protoc-gen-go
	go get -u google.golang.org/grpc/cmd/protoc-gen-go-grpc

generate:
	mkdir -p pkg/swagger
	make generate-users-api
	make generate-auth-api
	make generate-access-api
	$(LOCAL_BIN)/statik -src=pkg/swagger/ -include='*.css,*.html,*.js,*.json,*.png'


generate-users-api:
	mkdir -p pkg/users_v1 pkg/swagger
	protoc \
		--proto_path=api/users_v1 \
		--proto_path=vendor.protogen \
		--go_out=pkg/users_v1 --go_opt=paths=source_relative \
		--go-grpc_out=pkg/users_v1 --go-grpc_opt=paths=source_relative \
		--validate_out=lang=go:pkg/users_v1 --validate_opt=paths=source_relative \
		--grpc-gateway_out=pkg/users_v1 --grpc-gateway_opt=paths=source_relative \
		--openapiv2_out=allow_merge=true,merge_file_name=api:pkg/swagger \
		--plugin=protoc-gen-go=bin/protoc-gen-go.exe \
		--plugin=protoc-gen-go-grpc=bin/protoc-gen-go-grpc.exe \
		--plugin=protoc-gen-validate=bin/protoc-gen-validate.exe \
		--plugin=protoc-gen-grpc-gateway=bin/protoc-gen-grpc-gateway.exe \
		--plugin=protoc-gen-openapiv2=bin/protoc-gen-openapiv2.exe \
		api/users_v1/users.proto

generate-auth-api:
	mkdir -p pkg/auth_v1
	protoc --proto_path api/auth_v1 \
	--go_out=pkg/auth_v1 --go_opt=paths=source_relative \
	--plugin=protoc-gen-go=bin/protoc-gen-go.exe \
	--go-grpc_out=pkg/auth_v1 --go-grpc_opt=paths=source_relative \
	--plugin=protoc-gen-go-grpc=bin/protoc-gen-go-grpc.exe \
	api/auth_v1/auth.proto

generate-access-api:
	mkdir -p pkg/access_v1
	protoc --proto_path api/access_v1 \
	--go_out=pkg/access_v1 --go_opt=paths=source_relative \
	--plugin=protoc-gen-go=bin/protoc-gen-go.exe \
	--go-grpc_out=pkg/access_v1 --go-grpc_opt=paths=source_relative \
	--plugin=protoc-gen-go-grpc=bin/protoc-gen-go-grpc.exe \
	api/access_v1/access.proto
	
local-migration-status:
	$(LOCAL_BIN)/goose -dir $(LOCAL_MIGRATION_DIR) postgres "$(LOCAL_MIGRATION_DSN)" status -v

local-migration-up:
	$(LOCAL_BIN)/goose -dir $(LOCAL_MIGRATION_DIR) postgres "$(LOCAL_MIGRATION_DSN)" up -v

local-migration-down:
	$(LOCAL_BIN)/goose -dir $(LOCAL_MIGRATION_DIR) postgres "$(LOCAL_MIGRATION_DSN)" down -v


test:
	go clean -testcache
	go test ./... -covermode count -coverpkg=auth-service/internal/service/...,auth-service/internal/api/... -count 5


test-coverage:
	go clean -testcache
	go test ./... -coverprofile=coverage.tmp.out -covermode count -coverpkg=auth-service/internal/service/...,auth-service/internal/api/... -count 5
	grep -v 'mocks\|config' coverage.tmp.out  > coverage.out
	rm coverage.tmp.out
	go tool cover -html=coverage.out;
	go tool cover -func=./coverage.out | grep "total";
	grep -sqFx "/coverage.out" .gitignore || echo "/coverage.out" >> .gitignore

vendor-proto:
		@if [ ! -d vendor.protogen/validate ]; then \
			mkdir -p vendor.protogen/validate &&\
			git clone https://github.com/envoyproxy/protoc-gen-validate vendor.protogen/protoc-gen-validate &&\
			mv vendor.protogen/protoc-gen-validate/validate/*.proto vendor.protogen/validate &&\
			rm -rf vendor.protogen/protoc-gen-validate ;\
		fi
		@if [ ! -d vendor.protogen/google ]; then \
			git clone https://github.com/googleapis/googleapis vendor.protogen/googleapis &&\
			mkdir -p  vendor.protogen/google/ &&\
			mv vendor.protogen/googleapis/google/api vendor.protogen/google &&\
			rm -rf vendor.protogen/googleapis ;\
		fi
		@if [ ! -d vendor.protogen/protoc-gen-openapiv2 ]; then \
			mkdir -p vendor.protogen/protoc-gen-openapiv2/options &&\
			git clone https://github.com/grpc-ecosystem/grpc-gateway vendor.protogen/openapiv2 &&\
			mv vendor.protogen/openapiv2/protoc-gen-openapiv2/options/*.proto vendor.protogen/protoc-gen-openapiv2/options &&\
			rm -rf vendor.protogen/openapiv2 ;\
		fi


load-test:
	ghz \
		--proto api/users_v1/users.proto \
		--import-paths="vendor.protogen" \
        --call users_v1.UsersV1/Get \
		--data '{"id": 1}' \
		--rps 100 \
		--total 3000 \
		--insecure \
		localhost:9000

error-test:
	ghz \
		--proto api/users_v1/users.proto \
		--import-paths="vendor.protogen" \
        --call users_v1.UsersV1/Get \
		--data '{"id": 0}' \
		--rps 100 \
		--total 3000 \
		--insecure \
		localhost:9000


copy-to-server:
	scp -r migrations root@87.228.39.226:~
	scp -r prod.docker-compose.yml root@87.228.39.226:~
	scp -r prodMigration.Dockerfile root@87.228.39.226:~
	scp -r prodMigration.sh root@87.228.39.226:~
	scp -r prod.env root@87.228.39.226:~
	ssh root@87.228.39.226 "mv ~/prod.env ~/.env && chmod 600 ~/.env"
	scp -r metrics root@87.228.39.226:~
