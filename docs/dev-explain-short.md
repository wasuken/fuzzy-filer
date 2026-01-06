# fuzzy-filer 最低限理解ドキュメント

## 1. ファイル構成（5秒で理解）

```
main.go      → エントリーポイント、TUI制御、/dev/tty処理
model.go     → 状態管理、入力ハンドリング
scanner.go   → ファイル走査（filepath.WalkDir）
ranker.go    → スコアリング・ランキング
config.go    → 設定ファイル読み込み
keymap.go    → キーバインド定義
```

**役割分担**:
- **main.go**: ターミナル制御と入出力
- **model.go**: アプリのコア状態とロジック
- **scanner/ranker**: データ処理
- **config/keymap**: 設定

---

## 2. データフロー（30秒で理解）

```
起動
 ↓
LoadConfig() → デフォルトor~/.config/fuzzy-filer/config.json読む
 ↓
ScanFiles() → ディレクトリ走査、FileEntry[]作成
 ↓
【メインループ】
  ユーザー入力 (/dev/tty から読む)
   ↓
  HandleInput() → クエリ更新 or カーソル移動 or 選択
   ↓
  updateFilter() → RankEntries()でスコアリング
   ↓
  renderToTTY() → 画面再描画
 ↓
選択 → パスを標準出力に出して終了
```

**ポイント**:
- 入力: `/dev/tty`（パイプライン対応）
- 出力: `stdout`（パスのみ）
- 状態: `Model`構造体で全部持つ

---

## 3. 重要な設計判断

### 3.1 なぜ /dev/tty?

```go
// ❌ これだとパイプで死ぬ
reader := bufio.NewReader(os.Stdin)

// ✅ これならパイプでも動く
tty, _ := os.OpenFile("/dev/tty", os.O_RDWR, 0)
reader := bufio.NewReader(tty)
```

**理由**: `cat $(fuzzy-filer)` のとき、`stdin`はパイプになる。でもユーザー入力は必要。だから直接`/dev/tty`を開く。

---

### 3.2 raw mode とは

```go
setRawModeForFd(fd) // ターミナルを「生モード」に
```

**通常モード**:
- Enter押すまでバッファリング
- Ctrl+Cで即終了
- エコーバックあり

**rawモード**:
- 1文字ごとに即座に読める
- Ctrl+Cも普通の文字
- エコーバックなし（自分で表示制御）

**なぜ必要?**: `j/k`で即座にカーソル移動したいから

---

### 3.3 スコアリング優先順位

```go
// ranker.go の calculateScore()
ディレクトリ完全一致: 10000点
ファイル名前方一致:   1000点
ファイル名部分一致:    500点
親ディレクトリ一致:    100点
```

**設計思想**:
- ファイル名 > ディレクトリ名
- 前方一致 > 部分一致
- 下層ディレクトリ > 上層ディレクトリ

---

### 3.4 除外パターン

```go
// scanner.go
if shouldExclude(relPath, config.ExcludePatterns) {
    if d.IsDir() {
        return filepath.SkipDir  // ディレクトリごとスキップ
    }
    return nil  // ファイルだけスキップ
}
```

**ポイント**:
- `SkipDir`: そのディレクトリ配下を全部無視
- 深度チェック、ファイル数上限でも早期終了

---

## 4. カスタマイズポイント

### スコアリング変更したい

```go
// ranker.go の calculateScore() を編集
if strings.HasPrefix(nameLower, query) {
    score += 2000  // 前方一致の点数を2倍に
}
```

### キーバインド変更

```go
// keymap.go
Down: 'j',  // これを 'n' に変えるとか
```

### 除外パターン追加

```json
// ~/.config/fuzzy-filer/config.json
{
  "exclude_patterns": [
    "node_modules",
    "vendor",  // 追加
    "*.tmp"    // 追加
  ]
}
```

---

## 5. デバッグ方法

### ビルドして実行

```bash
go build -o fuzzy-filer
./fuzzy-filer ~/Documents
```

### 直接実行（開発中）

```bash
go run . ~/Documents
```

### ログ出す

```go
// エラーはstderrへ
fmt.Fprintf(os.Stderr, "Debug: %v\n", someValue)
```

---

## 6. よくあるトラブル

### 画面が壊れる

→ `restoreTerminalForFd()` が呼ばれてない  
→ `defer` の順序ミス  
→ 強制終了時は `reset` コマンド打つ

### パスが表示されない

→ `stdout` じゃなく `tty` に出力してる  
→ `fmt.Println()` は必ず最後（ターミナル復元後）

### パイプで動かない

→ `/dev/tty` 使ってない  
→ `os.Stdin` から読んでる

---

## 7. 拡張アイデア

実装したくなったら:

- **プレビュー機能**: 選択中のファイル内容を右側に表示
- **マルチセレクト**: Spaceで複数選択、パスを改行区切りで出力
- **正規表現**: クエリを正規表現で解釈
- **履歴**: 最近選んだファイルを優先表示
- **stdin対応**: `find | fuzzy-filer` で絞り込み

---

## 8. 10秒まとめ

```
ファイル走査 → ユーザー入力 → スコアリング → 表示 → 選択 → パス出力
```

- 入力: `/dev/tty`（パイプ対応）
- 状態: `Model`構造体
- 出力: `stdout`（パスのみ）
