# gin-sample-app

ブログ記事の投稿・取得を想定した Gin 製のサンプル Web API です。  
MVC（Model / Repository / Service / Handler）構成で、記事の作成・一覧・詳細・更新・削除の基本的な機能を提供します。

## 開発環境

- Go 1.24+（`go.mod` の toolchain 指定により自動取得）
- Gin v1.11
- 標準 `database/sql` + SQLite（modernc.org/sqlite）をデフォルト利用
- PostgreSQL（pgx/v5）にも接続可能
- 推奨: Docker / Make / golangci-lint / govulncheck（Makefileから実行）

## ファイル構成

```
gin-sample-app/
├── main.go                         # エントリーポイント。config読み込み・DI・ルーティング設定
├── cmd/
│   └── migrate/main.go             # migrate CLI（ランタイムでマイグレーション実行）
├── config/config.go                # 環境変数(APP_ENV, LOG_LEVEL, DB設定など)を読み込む
├── handler/
│   ├── admin_handler.go            # ログレベル管理API
│   └── post_handler.go             # POST CRUD HTTPハンドラ
├── internal/
│   ├── database/
│   │   ├── database.go             # DB接続ユーティリティ
│   │   ├── migrate.go              # 埋め込みマイグレーション適用機能
│   │   └── migrations/             # SQLite / Postgres 用マイグレーションSQL
│   ├── middleware/
│   │   ├── auth.go                 # APIキー認証
│   │   └── logging.go              # 構造化アクセスログ
│   └── server/server.go            # Ginサーバー組み立て
├── logger/
│   └── logger.go                   # Zapロガー初期化とランタイム制御
├── model/post.go                   # ドメインモデル
├── repository/
│   └── post_repository.go          # SQL / in-memory リポジトリ
├── service/post_service.go         # ビジネスロジック層
├── integration/                    # サービス+リポジトリの統合テスト
├── docs/
│   ├── openapi.yaml                # OpenAPI 3.0 定義
│   └── embed.go                    # 埋め込みヘルパー
├── Dockerfile                      # マルチステージビルド
├── Makefile                        # 開発用コマンド
├── .env.sample                     # 開発用設定サンプル
└── ...
```

## 環境変数

`.env` を作成すると `config.Load()` が自動で読み込みます。  
`make run` や `make docker-run` は `.env` を利用する前提です。

| 変数      | デフォルト    | 説明                       |
|-----------|---------------|----------------------------|
| `APP_ENV` | `dev`         | `dev` / `stg` / `prd`       |
| `PORT`    | `8080`        | HTTPサーバーの待受ポート   |
| `LOG_LEVEL` | `debug`     | Zapのログレベル（`debug` / `info` / `warn` / `error` など） |
| `API_KEY`   | *(空文字)*  | 設定すると更新系APIで `X-API-Key` ヘッダー必須 |
| `DB_DRIVER` | `sqlite`     | `sqlite` / `postgres` / `pgx` などドライバ名 |
| `DB_DSN`    | `file:tmp/app.db?_foreign_keys=1` | ドライバへ渡す接続文字列 |

### `.env` サンプル

```
APP_ENV=dev
PORT=8080
LOG_LEVEL=debug
DB_DRIVER=sqlite
DB_DSN=file:tmp/app.db?_foreign_keys=1
```

## 初期セットアップ

```bash
cp .env.sample .env
```

値を必要に応じて編集してから、以下の手順で起動してください。

## データベース

- デフォルトは SQLite（modernc.org/sqlite）です。`DB_DSN=file:tmp/app.db?_foreign_keys=1` により `tmp/app.db` が自動生成され、外部キー制約が有効になります。
- PostgreSQL を利用する場合は `.env` に `DB_DRIVER=postgres` と接続文字列（例: `DB_DSN=postgres://user:pass@localhost:5432/gin_sample?sslmode=disable`）を指定してください。
- アプリ起動時にマイグレーション (`golang-migrate/migrate`) を自動適用し、テーブルを最新の状態に更新します。

### マイグレーションの操作

- 最新化: `make migrate-up`
- 指定ステップ移動: `make migrate-steps STEPS=1`
- 全てロールバック: `make migrate-down`
- 静的検査: `make migrate-lint`

`cmd/migrate` は `.env` を読み込んだ上で `database/sql` を利用し、アプリと同じ接続設定でマイグレーションを実行します。

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
| GET      | `/admin/log-level` | 現在のログレベルを取得（APIキー必須） |
| PUT      | `/admin/log-level` | ログレベルを更新（APIキー必須） |

## OpenAPI / API スキーマ共有

- 仕様書は `docs/openapi.yaml` として管理し、アプリ起動中は `GET /openapi.yaml` でダウンロードできます。
- ブラウザから `http://localhost:8080/docs/swagger`（Swagger UI）や `http://localhost:8080/docs/redoc`（ReDoc）にアクセスすると、組み込みビューアでスキーマを閲覧できます。
- ローカルで独自に Swagger UI を起動する場合は `npx swagger-ui-watcher docs/openapi.yaml` も使えます。
- スキーマを更新した場合は Pull Request に `docs/openapi.yaml` の差分が含まれるよう注意してください。
- Spectral による lint は `.spectral.yaml`（`extends: spectral:oas`）を参照して実行されます。

## テストと品質管理

| コマンド        | 内容                                  |
|-----------------|---------------------------------------|
| `make test`     | `go test -v ./...`                    |
| `make lint`     | `go vet ./...`                        |
| `make vuln`     | `govulncheck ./...`（未インストール時は go install） |
| `make migrate-up` | マイグレーションを最新まで適用 |
| `make migrate-down` | マイグレーションを全てロールバック |
| `make migrate-lint` | マイグレーションファイル構成を検証 |
| `make openapi-lint` | OpenAPI スキーマを lint（Spectral） |

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

- `LOG_LEVEL` で Zap のログレベル（`debug` / `info` / `warn` / `error` など）を制御できます。デフォルトは `debug`。
- `APP_ENV=prd` では JSON 形式の構造化ログを出力し、それ以外の環境では開発向けのカラー表示を行います。
- すべてのログには `timestamp` / `level` / `message` に加えて `env` と `service`（固定値: `gin-sample-app`）が付与されます。
- Gin のリクエストログは `status`, `latency`, `client_ip`, `user_agent` などのフィールドを含む構造化ログとして記録されます。
- ランタイムでは `/admin/log-level` にアクセスすることでレベルを取得・更新できます。例：
  - 取得: `curl -H "X-API-Key: your-api-key" http://localhost:8080/admin/log-level`
  - 更新: `curl -X PUT -H "Content-Type: application/json" -H "X-API-Key: your-api-key" -d '{"level":"info"}' http://localhost:8080/admin/log-level`
  - PUT リクエストが成功すると、旧レベル・新レベル・リクエスト送信元IPなどが Zap の Info ログとして監査出力されます。

## 今後の発展例

- 認証・認可や中間層のミドルウェア追加
