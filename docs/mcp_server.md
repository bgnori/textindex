# MCP サーバー

`textindex mcp` は読み込み済みインデックスを Model Context Protocol (MCP) の stdio サーバーとして公開します。MCP クライアントから検索とインデックス情報取得を実行できます。

## 起動

```bash
textindex mcp --index data/moby_dick.idx
```

MCP クライアントはサーバープロセスを起動し、stdin/stdout で通信します。標準出力はプロトコル用なので、ターミナルへ直接起動したときに通常の CLI 出力は表示されません。`--index` を省略すると `data/moby_dick.idx` を読み込みます。

最低 Go バージョンは公式 MCP Go SDK に合わせて Go 1.23 です。

## 提供ツール

### `search`

入力:

- `pattern`（必須）: 検索パターン
- `limit`（任意、既定 20、1〜1000）: 最大結果数
- `context`（任意、既定 80、0〜1000）: マッチ前後のコンテキスト長（バイト数）。0 を指定するとマッチ部分だけを返します。

返却値はパターン、各マッチのバイト位置とコンテキスト付きスニペット、結果が上限で切り詰められたかどうかです。FM-index は星なし正規表現、Suffix Array と双方向 FM-index はリテラル検索として解釈します。

### `index_info`

読み込み中のインデックスファイル、バックエンド種別、テキスト長を返します。FM-index の場合は suffix-array 長とアルファベットサイズも含みます。

## MCP クライアント設定例

```json
{
  "mcpServers": {
    "textindex": {
      "command": "/absolute/path/to/textindex",
      "args": ["mcp", "--index", "/absolute/path/to/data/moby_dick.idx"]
    }
  }
}
```

`textindex` は先に `go build -o textindex ./cmd/textindex` などでビルドしてください。クライアントに登録後、`search` と `index_info` を利用できます。