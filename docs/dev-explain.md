# Go開発者必修 コードブロック別ランキング

## 【1位】main.go - システムプログラミングの本質

### S級ブロック（絶対理解必須）

#### 🔥 ブロック1-1: /dev/tty を開く（最重要）

```go
// /dev/ttyを開く（パイプライン対応）♧
tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
if err != nil {
    fmt.Fprintf(os.Stderr, "Error opening /dev/tty: %v\n", err)
    os.Exit(1)
}
defer tty.Close()
```

**なぜ重要**: 
- **パイプライン対応の核心技術**
- `stdin`がパイプでもユーザー入力を受け付ける唯一の方法
- fzf, peco等のCLIツールが必ず使う定石

**学び**:
- `os.OpenFile(path, flag, perm)` の使い方
- `/dev/tty` = 制御端末への直接アクセス
- `O_RDWR` = 読み書き両方

---

#### 🔥 ブロック1-2: raw mode設定（超重要）

```go
// setRawModeForFd は指定fdをrawモードに設定♥
func setRawModeForFd(fd int) (*syscall.Termios, error) {
	oldState := &syscall.Termios{}

	if _, _, err := syscall.Syscall6(syscall.SYS_IOCTL, uintptr(fd),
		uintptr(syscall.TCGETS), uintptr(unsafe.Pointer(oldState)),
		0, 0, 0); err != 0 {
		return nil, err
	}

	newState := *oldState
	newState.Lflag &^= syscall.ECHO | syscall.ICANON | syscall.ISIG
	newState.Iflag &^= syscall.IXON | syscall.ICRNL
	newState.Cc[syscall.VMIN] = 1
	newState.Cc[syscall.VTIME] = 0

	if _, _, err := syscall.Syscall6(syscall.SYS_IOCTL, uintptr(fd),
		uintptr(syscall.TCSETS), uintptr(unsafe.Pointer(&newState)),
		0, 0, 0); err != 0 {
		return nil, err
	}

	return oldState, nil
}
```

**なぜ重要**:
- **ターミナルプログラミングの基礎**
- vim, less等のTUIアプリが必ず使う
- syscall直叩きの典型例

**学び**:
- `syscall.Syscall6()` - システムコール直接実行
- `TIOCGWINSZ/TCGETS/TCSETS` - ターミナル制御ioctl
- `unsafe.Pointer` - Goでのメモリ直接操作
- `&^=` - ビットクリア演算子（Go特有）

**各フラグの意味**:
- `ECHO` off: 入力文字を自動表示しない
- `ICANON` off: 行バッファリング無効（1文字ずつ読む）
- `ISIG` off: Ctrl+Cを特殊扱いしない
- `IXON` off: Ctrl+S/Qのフロー制御無効
- `VMIN=1, VTIME=0`: 1文字でも即座に返す

---

#### 🔥 ブロック1-3: メインループ（重要）

```go
// メインループ♠
reader := bufio.NewReader(tty)
var selectedPath string

for {
    r, _, err := reader.ReadRune()
    if err != nil {
        break
    }

    // 入力処理♧
    quit, path, err := model.HandleInput(r)
    if err != nil {
        fmt.Fprintf(tty, "\n\033[1;31mError: %v\033[0m\n", err)
        continue
    }

    if quit {
        selectedPath = path
        break
    }

    // 再描画♠
    renderToTTY(model, tty)
}
```

**なぜ重要**:
- **イベント駆動プログラミングの基本形**
- TUI アプリの標準パターン

**学び**:
- `bufio.Reader.ReadRune()` - UTF-8対応の文字読み込み
- ループ+イベントハンドラのパターン
- エラーハンドリング（continue vs break）

---

#### 🔥 ブロック1-4: クリーンアップ順序（重要）

```go
// 終了時クリーンアップ♧
// 1. 画面クリア
fmt.Fprint(tty, "\033[2J\033[H")

// 2. ターミナル状態を復元
restoreTerminalForFd(int(tty.Fd()), oldState)

// 3. カーソル表示
fmt.Fprint(tty, "\033[?25h")

// 4. パスを標準出力に出力（ttyではなくstdout）♥
if selectedPath != "" {
    fmt.Println(selectedPath)
}
```

**なぜ重要**:
- **リソース解放の順序が生死を分ける**
- defer の落とし穴を回避

**学び**:
- クリーンアップは明示的に順序制御
- `tty` と `stdout` の使い分け
- ANSIエスケープシーケンス（`\033[2J` = 画面クリア等）

---

### A級ブロック（知っとくべき）

#### ブロック1-5: エスケープシーケンス

```go
fmt.Fprint(tty, "\033[2J\033[H\033[?25l")
//               ^^^^^^ 画面クリア
//                     ^^^^^^ カーソルを左上に
//                           ^^^^^^^^ カーソル非表示
```

**学び**:
- `\033[` = CSI（Control Sequence Introducer）
- `2J` = 画面全体クリア
- `H` = カーソルホーム位置
- `?25l` / `?25h` = カーソル非表示/表示

---

## 【2位】model.go - 状態管理とイベント処理

### S級ブロック

#### 🔥 ブロック2-1: Model構造体（超重要）

```go
// Model はアプリケーション状態♠
type Model struct {
	currentDir      string
	allEntries      []FileEntry
	filteredEntries []FileEntry
	query           string
	cursor          int
	keymap          KeyMap
	config          Config
	width           int
	height          int
}
```

**なぜ重要**:
- **状態管理の設計思想**
- 全ての状態を1箇所に集約

**学び**:
- 構造体ベースの状態管理（React風）
- プライベートフィールド（小文字始まり）
- 責務の分離（表示状態、データ、設定）

---

#### 🔥 ブロック2-2: HandleInput - イベントハンドラ（最重要）

```go
func (m *Model) HandleInput(r rune) (bool, string, error) {
	switch {
	case r == m.keymap.Quit:
		return true, "", nil

	case r == m.keymap.Down:
		if m.cursor < len(m.filteredEntries)-1 {
			m.cursor++
		}

	case r == m.keymap.Up:
		if m.cursor > 0 {
			m.cursor--
		}

	case r == m.keymap.Enter:
		if len(m.filteredEntries) > 0 {
			selected := m.filteredEntries[m.cursor]
			if selected.IsDir {
				return false, "", m.changeDirectory(selected.Path)
			}
			fullPath := filepath.Join(m.currentDir, selected.Path)
			return true, fullPath, nil
		}

	case r == m.keymap.Backspace || r == m.keymap.DeleteQuery:
		if len(m.query) > 0 {
			m.query = m.query[:len(m.query)-1]
			m.updateFilter()
		}

	default:
		if r >= 32 && r < 127 {
			m.query += string(r)
			m.updateFilter()
		}
	}

	return false, "", nil
}
```

**なぜ重要**:
- **イベント駆動の典型実装**
- 状態更新パターンの教科書

**学び**:
- `switch` に条件式（Go特有、他言語の `if-else if` 相当）
- メソッドレシーバー `(m *Model)` - 構造体のメソッド
- 多値return `(bool, string, error)` - Goの慣用句
- 境界チェック `if m.cursor < len(...)-1`
- 文字列結合 `m.query += string(r)`

---

#### 🔥 ブロック2-3: NewModel - コンストラクタパターン（重要）

```go
func NewModel(startDir string) (*Model, error) {
	absDir, err := filepath.Abs(startDir)
	if err != nil {
		return nil, err
	}

	config := LoadConfig()

	entries, err := ScanFiles(absDir, config)
	if err != nil {
		return nil, err
	}

	width, height := getTerminalSize()

	m := &Model{
		currentDir:      absDir,
		allEntries:      entries,
		filteredEntries: RankEntries(entries, ""),
		query:           "",
		cursor:          0,
		keymap:          DefaultKeyMap(),
		config:          config,
		width:           width,
		height:          height,
	}

	return m, nil
}
```

**なぜ重要**:
- **Goのコンストラクタパターン**（New〇〇関数）
- 初期化の責務分離

**学び**:
- `*Model` を返す（ポインタ返却）
- エラーチェーン `if err != nil { return nil, err }`
- 構造体リテラル初期化
- 複合的な初期化処理の集約

---

### A級ブロック

#### ブロック2-4: getTerminalSize - syscallの実用例

```go
func getTerminalSize() (int, int) {
	type winsize struct {
		Row    uint16
		Col    uint16
		Xpixel uint16
		Ypixel uint16
	}

	ws := &winsize{}
	retCode, _, _ := syscall.Syscall(syscall.SYS_IOCTL,
		uintptr(syscall.Stdout),
		uintptr(syscall.TIOCGWINSZ),
		uintptr(unsafe.Pointer(ws)))

	if int(retCode) == -1 {
		return 80, 24
	}
	return int(ws.Col), int(ws.Row)
}
```

**学び**:
- 内部構造体定義（C構造体のマッピング）
- `TIOCGWINSZ` - ウィンドウサイズ取得ioctl
- デフォルト値パターン（80x24）

---

#### ブロック2-5: View - レンダリングロジック

```go
func (m *Model) View() string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("\033[1;36m%s\033[0m ", m.currentDir))
	b.WriteString(fmt.Sprintf("\033[2m[%d files]\033[0m\n", len(m.allEntries)))
	b.WriteString(fmt.Sprintf("> %s\033[K\n", m.query))
	b.WriteString(strings.Repeat("─", min(m.width, 80)) + "\n")

	for i, entry := range m.filteredEntries {
		cursor := "  "
		if i == m.cursor {
			cursor = "\033[1;33m>\033[0m "
		}

		icon := " "
		color := "\033[0m"
		if entry.IsDir {
			icon = "📁"
			color = "\033[1;34m"
		} else {
			icon = "📄"
		}

		displayPath := entry.Name
		if entry.DirPath != "." {
			displayPath = filepath.Join(entry.DirPath, entry.Name)
		}

		b.WriteString(fmt.Sprintf("%s %s %s%s\033[0m\n", 
			cursor, icon, color, displayPath))
	}

	b.WriteString("\n")
	b.WriteString("\033[2m[j/k]移動 [Enter]選択 [q]終了\033[0m")

	return b.String()
}
```

**学び**:
- `strings.Builder` - 効率的な文字列結合
- ANSIカラーコード（`\033[1;36m` = シアン等）
- Unicode絵文字の使用
- 条件分岐での表示切り替え

---

## 【3位】scanner.go - ファイル走査の定石

### S級ブロック

#### 🔥 ブロック3-1: ScanFiles - WalkDirの使い方（最重要）

```go
func ScanFiles(rootDir string, config Config) ([]FileEntry, error) {
	var entries []FileEntry
	fileCount := 0

	err := filepath.WalkDir(rootDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // エラーは無視して継続
		}

		if path == rootDir {
			return nil
		}

		relPath, _ := filepath.Rel(rootDir, path)

		// 深度チェック♠
		depth := strings.Count(relPath, string(os.PathSeparator)) + 1
		if depth > config.MaxDepth {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// 除外パターンチェック♧
		if shouldExclude(relPath, config.ExcludePatterns) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// 隠しファイル/ディレクトリはスキップ♥
		if strings.HasPrefix(d.Name(), ".") {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// ファイル数上限チェック♠
		fileCount++
		if fileCount > config.MaxFiles {
			return filepath.SkipAll
		}

		dirPath := filepath.Dir(relPath)

		entries = append(entries, FileEntry{
			Path:    relPath,
			Name:    d.Name(),
			IsDir:   d.IsDir(),
			DirPath: dirPath,
		})

		return nil
	})

	return entries, err
}
```

**なぜ重要**:
- **ファイル走査の完全版テンプレート**
- 早期リターンによる効率化
- 実用的なエラーハンドリング

**学び**:
- `filepath.WalkDir()` - ディレクトリ再帰走査
- `filepath.SkipDir` - サブディレクトリをスキップ
- `filepath.SkipAll` - 走査全体を中断
- `filepath.Rel()` - 相対パス計算
- `os.PathSeparator` - OS依存のパス区切り文字
- クロージャー内でのエラー制御

**早期リターンの威力**:
```go
if d.IsDir() {
    return filepath.SkipDir  // ディレクトリ全体をスキップ
}
```
→ `node_modules`配下の数万ファイルを一瞬で飛ばせる

---

### A級ブロック

#### ブロック3-2: FileEntry構造体

```go
type FileEntry struct {
	Path    string
	Name    string
	IsDir   bool
	DirPath string
}
```

**学び**:
- シンプルなデータ構造
- 必要最小限のフィールド

---

## 【4位】config.go - 設定管理の標準形

### S級ブロック

#### 🔥 ブロック4-1: LoadConfig - 設定読み込みパターン（重要）

```go
func LoadConfig() Config {
	configPath := getConfigPath()
	
	data, err := os.ReadFile(configPath)
	if err != nil {
		return DefaultConfig()
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return DefaultConfig()
	}

	return config
}
```

**なぜ重要**:
- **設定ファイル読み込みの定石**
- フォールバック処理の実装例

**学び**:
- `os.ReadFile()` - ファイル一括読み込み
- `json.Unmarshal()` - JSON→構造体
- エラー時のデフォルト値返却パターン

---

#### 🔥 ブロック4-2: shouldExclude - パターンマッチング（重要）

```go
func shouldExclude(path string, patterns []string) bool {
	for _, pattern := range patterns {
		if strings.HasPrefix(pattern, "*.") {
			ext := pattern[1:]
			if strings.HasSuffix(path, ext) {
				return true
			}
		} else {
			if strings.Contains(path, pattern) {
				return true
			}
		}
	}
	return false
}
```

**学び**:
- シンプルなワイルドカード実装
- `range` でのスライス走査
- 早期リターン（見つかったら即終了）

---

### A級ブロック

#### ブロック4-3: getConfigPath

```go
func getConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".fuzzy-filer.json"
	}
	return filepath.Join(home, ".config", "fuzzy-filer", "config.json")
}
```

**学び**:
- `os.UserHomeDir()` - ホームディレクトリ取得
- `filepath.Join()` - OS依存のパス結合
- フォールバックパターン

---

## 【5位】ranker.go - アルゴリズム実装例

### A級ブロック

#### ブロック5-1: RankEntries - ソート実装（重要）

```go
func RankEntries(entries []FileEntry, query string) []FileEntry {
	if query == "" {
		return entries[:min(10, len(entries))]
	}

	query = strings.ToLower(query)
	var scored []ScoredEntry

	for _, entry := range entries {
		score := calculateScore(entry, query)
		if score > 0 {
			scored = append(scored, ScoredEntry{
				Entry: entry,
				Score: score,
			})
		}
	}

	sort.Slice(scored, func(i, j int) bool {
		if scored[i].Score != scored[j].Score {
			return scored[i].Score > scored[j].Score
		}
		if scored[i].Entry.IsDir != scored[j].Entry.IsDir {
			return scored[i].Entry.IsDir
		}
		return scored[i].Entry.Name < scored[j].Entry.Name
	})

	result := make([]FileEntry, 0, min(10, len(scored)))
	for i := 0; i < min(10, len(scored)); i++ {
		result = append(result, scored[i].Entry)
	}

	return result
}
```

**学び**:
- `sort.Slice()` - カスタムソート
- 多段階ソート（スコア→種類→名前）
- スライスの事前容量確保 `make([]T, 0, capacity)`
- 早期リターン（クエリ空文字列）

---

#### ブロック5-2: calculateScore - スコアリングロジック

```go
func calculateScore(entry FileEntry, query string) int {
	nameLower := strings.ToLower(entry.Name)
	score := 0

	if entry.IsDir && nameLower == query {
		return 10000
	}

	if strings.HasPrefix(nameLower, query) {
		score += 1000
		if entry.IsDir {
			score += 500
		}
		return score
	}

	if idx := strings.Index(nameLower, query); idx >= 0 {
		score += 500 - idx*10
		if entry.IsDir {
			score += 200
		}
		return score
	}

	if entry.DirPath != "." {
		dirs := strings.Split(entry.DirPath, string(filepath.Separator))
		for i := len(dirs) - 1; i >= 0; i-- {
			dirLower := strings.ToLower(dirs[i])
			if strings.Contains(dirLower, query) {
				score += 100 - (len(dirs)-1-i)*20
				break
			}
		}
	}

	return score
}
```

**学び**:
- 段階的スコアリング設計
- `strings.Index()` - 部分文字列位置取得
- `strings.Split()` - 文字列分割
- 逆順ループ `for i := len(x)-1; i >= 0; i--`

---

# 総まとめ: Go開発で絶対覚えるべきTOP10

1. **syscall直叩き** (`main.go` raw mode)
2. **/dev/tty制御** (`main.go` OpenFile)
3. **構造体ベース状態管理** (`model.go` Model)
4. **メソッドレシーバー** (`model.go` HandleInput)
5. **filepath.WalkDir** (`scanner.go` ScanFiles)
6. **早期リターン** (`scanner.go` SkipDir)
7. **json.Unmarshal** (`config.go` LoadConfig)
8. **sort.Slice** (`ranker.go` RankEntries)
9. **strings.Builder** (`model.go` View)
10. **多値return + エラーハンドリング** (全ファイル)
