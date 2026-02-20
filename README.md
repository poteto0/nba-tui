# nba-tui

NBAのスコアボードを表示するTUIクライアントです。

## 特徴

- **リアルタイム更新**: 30秒ごとに自動でスコアを更新します。
- **直感的な操作**: `vim`に近いキーバインドで操作可能です。
- **レスポンシブデザイン**: ターミナルのサイズに合わせてレイアウト（列数）を自動調整します。
- **ブラウザ連携**: 選択中の試合を `ctrl+w` でNBA.comの試合ページとして開けます。
- **視認性**:
  - 選択中の試合は緑色の枠線で強調されます。
  - 勝利チームのチーム名とスコアは太字で表示されます。
  - スコアは桁数に合わせて適切にパディングされ、見やすく配置されます。
  - 最終更新日時が画面上部に表示されます。

## インストール・実行方法

### 実行

```bash
go run cmd/nba-tui/main.go
```

### ビルド

```bash
go build -o nba-tui cmd/nba-tui/main.go
./nba-tui
```

## キーバインド

- `h`, `left`: 左の試合へ移動
- `l`, `right`: 右の試合へ移動
- `k`, `up`: 上の試合へ移動（グリッド表示時）
- `j`, `down`: 下の試合へ移動（グリッド表示時）
- `ctrl+w`: 選択中の試合をブラウザで開く
- `q`, `esc`, `ctrl+c`: 終了

## 開発

### テストの実行

```bash
go test ./internal/ui/scoreboard/...
```

### 使用ライブラリ

- [bubbletea](https://github.com/charmbracelet/bubbletea)
- [lipgloss](https://github.com/charmbracelet/lipgloss)
- [go-nba-sdk](https://github.com/poteto0/go-nba-sdk)
