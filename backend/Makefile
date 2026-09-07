.DEFAULT_GOAL := help

# 開発環境ではbackend/.envを読み込み、各コマンドへ環境変数として渡す。
# .envがないCI環境でもtestなどを実行できるよう、includeは任意にしている。
-include .env
export

# Supabaseのtransaction poolerではnamed prepared statementを共有できないため、
# Gooseにはstatement cacheを使わないpgxのexec modeを指定する。
DATABASE_URL_QUERY_SEPARATOR := $(if $(findstring ?,$(DATABASE_URL)),&,?)
export GOOSE_DRIVER := postgres
export GOOSE_DBSTRING := $(DATABASE_URL)$(DATABASE_URL_QUERY_SEPARATOR)default_query_exec_mode=exec
export GOOSE_MIGRATION_DIR := migrations

.PHONY: help dev build test test-integration test-race vet lint sqlc-vet fmt fmt-check check tidy generate generate-rpc generate-sql migrate-up migrate-down migrate-status migrate-create

help: ## 利用できるtargetを表示する
	@awk 'BEGIN {FS = ":.*## "} /^[a-zA-Z0-9_-]+:.*## / {printf "%-18s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

dev: ## API serverを起動する（.envを読み込む）
	go run ./cmd/server

build: ## API serverをbin/serverへbuildする
	mkdir -p bin
	go build -o bin/server ./cmd/server

test: ## unit testを実行する
	go test ./...

test-integration: export TEST_DATABASE_URL := $(DATABASE_URL)
test-integration: ## PostgreSQLを使うintegration testを実行する
	go test -count=1 ./internal/modules/account/internal/adapter/postgres

test-race: ## race detector付きでtestを実行する
	go test -race ./...

vet: ## go vetを実行する
	go vet ./...

lint: ## golangci-lintを実行する
	golangci-lint run ./...

sqlc-vet: ## sqlcのschemaとquery定義を検証する
	sqlc vet

fmt: ## 生成コード以外のGoコードをformatする
	gofmt -w $$(find . -type f -name '*.go' -not -path './generated/*')

fmt-check: ## formatされていないGoコードがないか確認する
	@files="$$(gofmt -l $$(find . -type f -name '*.go' -not -path './generated/*'))"; \
	if [ -n "$$files" ]; then echo "formatが必要なファイル:"; echo "$$files"; exit 1; fi

check: fmt-check vet test lint sqlc-vet ## format・静的解析・testをまとめて実行する

tidy: ## go.modとgo.sumを整理する
	go mod tidy

generate: generate-rpc generate-sql ## RPC・DBアクセスコードを再生成する

generate-rpc: ## repository rootのbuf設定からRPCコードを再生成する
	cd .. && buf generate

generate-sql: ## sqlcで型安全なDBアクセスコードを再生成する
	sqlc generate

migrate-up: ## 未適用のDB migrationをすべて適用する
	goose up

migrate-down: ## 最新のDB migrationを1つ戻す
	goose down

migrate-status: ## DB migrationの適用状況を表示する
	goose status

migrate-create: ## migrationを作成する（例: make migrate-create name=add_reviews）
	@if [ -z "$(name)" ]; then echo "nameを指定してください"; exit 1; fi
	goose -dir migrations create "$(name)" sql
