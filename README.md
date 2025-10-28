# gin-sample-app

ブログ記事の投稿・取得を想定した Gin 製のサンプル Web API です。  
MVC（Model / Repository / Service / Handler）構成で、記事の作成・一覧・詳細・更新・削除の基本的な機能を提供します。

## 開発環境

- Go 1.23+（`go.mod` の toolchain 指定により自動取得）
- Gin v1.11
- GORM は未使用（インメモリのリポジトリ実装）
- 推奨: Docker / Make / golangci-lint / govulncheck（Makefileから実行）

## ファイル構成

```
gin-sample-app/
├── main.go                # エントリーポイント。config読み込み・DI・ルーティング設定
├── config/config.go       # 環境変数(APP_ENV, PORT, LOG_LEVEL)を読み込む
├── handler/post_handler.go
├── service/post_service.go
├── repository/post_repository.go
├── model/post.go
├── internal/middleware/   # Gin用ミドルウェア（認証・ログ）
├── integration/           # サービス＋リポジトリの統合テスト
├── Dockerfile             # マルチステージビルド
├── .env.sample            # 開発用設定サンプル
└── ...
```

## 環境変数

`.env` を作成すると `config.Load()` が自動で読み込みます。  
`make run` や `make docker-run` は `.env` を利用する前提です。

| 変数      | デフォルト    | 説明                       |
|-----------|---------------|----------------------------|
| `APP_ENV` | `dev`         | `dev` / `stg` / `prd`       |
| `PORT`    | `8080`        | HTTPサーバーの待受ポート   |
| `LOG_LEVEL` | `debug`     | 将来的なログレベル設定用   |
| `API_KEY`   | *(空文字)*  | 設定すると更新系APIで `X-API-Key` ヘッダー必須 |

### `.env` サンプル

```
APP_ENV=dev
PORT=8080
LOG_LEVEL=debug
```

## 初期セットアップ

```bash
cp .env.sample .env
```

値を必要に応じて編集してから、以下の手順で起動してください。

## 実行方法

実行方式は **Goで直接起動** と **Dockerコンテナで起動** のどちらかを選択してください。

### 1. Goで直接起動

```bash
go run main.go
# または
make run              # godotenvが自動で .env を読み込む
```

#### Air を使ったホットリロード（任意）

```bash
go install github.com/air-verse/air@latest
air
```

`.air.toml` では `main.go` の変更を監視し、`tmp/main` バイナリで再起動する設定です。

### 2. Dockerコンテナで起動

```bash
make docker-build           # イメージ作成（デフォルト: gin-sample-app:latest）
make docker-run             # 8080番ポートで起動（--env-file .env）
```

停止やイメージ削除:
```bash
Ctrl+C                        # make docker-run を中断
make docker-clean             # Dockerイメージ削除
```

## API エンドポイント

| メソッド | パス           | 説明               |
|----------|----------------|--------------------|
| POST     | `/posts`       | 記事の新規作成     |
| GET      | `/posts`       | 記事一覧を取得     |
| GET      | `/posts/:id`   | 記事の詳細を取得   |
| PATCH    | `/posts/:id`   | 記事の部分更新     |
| DELETE   | `/posts/:id`   | 記事の削除         |

### 例: 記事作成

```bash
curl -X POST http://localhost:8080/posts \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key" \
  -d '{"title":"Gin入門","content":"本文","author":"Alice"}'
```

> `API_KEY` を設定している場合は `X-API-Key` ヘッダーを忘れずに付与してください。

### 例: 記事の部分更新（タイトル変更）

```bash
curl -X PATCH http://localhost:8080/posts/1 \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key" \
  -d '{"title":"新しいタイトル"}'
```

### 例: 記事の削除

```bash
curl -X DELETE http://localhost:8080/posts/1 \
  -H "X-API-Key: your-api-key"
```

## テストと品質管理

| コマンド        | 内容                                  |
|-----------------|---------------------------------------|
| `make test`     | `go test -v ./...`                    |
| `make lint`     | `go vet ./...`                        |
| `make vuln`     | `govulncheck ./...`（未インストール時は go install） |

### GitHub Actions CI

`.github/workflows/ci.yml` にて以下を自動実行します：
- `make lint`
- `make vuln`
- `make test`

## 依存管理

- Dependabot（`.github/dependabot.yml`）が Go Modules と GitHub Actions の更新を週次で確認し、PRを自動作成します。

## ディレクトリ別テスト

- `repository/`・`service/`・`handler/`・`integration/` にレイヤーごとのテストを用意し、責務ごとに検証しています。
- `integration/` は HTTP 層を含めず、Service + Repository を対象にした統合テストです。

## ログ

アプリ起動時に `APP_ENV` とリッスンポート、`LOG_LEVEL` を標準出力に出力します（例: `starting gin-sample-app env=dev listen=:8080 log_level=debug`）。

## 今後の発展例

- Repository を RDB（例: SQLite, PostgreSQL）実装へ差し替える
- Config / Logger を Zap + 構造化ログへ統合
- OpenAPI や Swagger を導入して API スキーマを共有
- 認証・認可や中間層のミドルウェア追加
