package main

import (
	"fmt"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"
)

// Model はアプリケーション状態
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
	previewCache    []string // プレビュー内容キャッシュ
}

// NewModel は新しいモデルを作成
func NewModel(startDir string) (*Model, error) {
	absDir, err := filepath.Abs(startDir)
	if err != nil {
		return nil, err
	}

	// 設定読み込み
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
		previewCache:    nil,
	}

	// 初期プレビュー生成
	m.updatePreview()

	return m, nil
}

// updateFilter はクエリに基づいてフィルタ更新
func (m *Model) updateFilter() {
	m.filteredEntries = RankEntries(m.allEntries, m.query)
	if m.cursor >= len(m.filteredEntries) {
		m.cursor = max(0, len(m.filteredEntries)-1)
	}
	m.updatePreview()
}

// updatePreview はプレビューを更新
func (m *Model) updatePreview() {
	if !m.config.EnablePreview || len(m.filteredEntries) == 0 {
		m.previewCache = nil
		return
	}

	selected := m.filteredEntries[m.cursor]
	fullPath := filepath.Join(m.currentDir, selected.Path)
	m.previewCache = GeneratePreview(fullPath, m.config.PreviewLines)
}

// changeDirectory はディレクトリ変更
func (m *Model) changeDirectory(newDir string) error {
	absDir := filepath.Join(m.currentDir, newDir)
	entries, err := ScanFiles(absDir, m.config)
	if err != nil {
		return err
	}

	m.currentDir = absDir
	m.allEntries = entries
	m.query = ""
	m.cursor = 0
	m.updateFilter()
	return nil
}

// model.go
func (m *Model) View() string {
	var b strings.Builder

	// ヘッダー
	b.WriteString(fmt.Sprintf("\033[1;36m%s\033[0m ", m.currentDir))
	b.WriteString(fmt.Sprintf("\033[2m[%d files]\033[0m\n", len(m.allEntries)))
	b.WriteString(fmt.Sprintf("> %s\033[K\n", m.query))
	b.WriteString(strings.Repeat("─", min(m.width, 80)) + "\n")

	// **プレビュー有効時は左右分割**♥
	if m.config.EnablePreview && len(m.previewCache) > 0 {
		return m.viewWithPreview()
	}

	// 通常表示（プレビューなし）
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
	b.WriteString("\033[2m[Ctrl+N/P]移動 [Enter]選択 [Ctrl+D]終了\033[0m")

	return b.String()
}

// viewWithPreview は左右分割プレビュー表示♠
// model.go
func (m *Model) viewWithPreview() string {
	var b strings.Builder

	// ヘッダー
	b.WriteString(fmt.Sprintf("\033[1;36m%s\033[0m ", m.currentDir))
	b.WriteString(fmt.Sprintf("\033[2m[%d files]\033[0m\n", len(m.allEntries)))
	b.WriteString(fmt.Sprintf("> %s\033[K\n", m.query))

	// 区切り線
	leftWidth := m.width / 2
	rightWidth := m.width - leftWidth - 1

	b.WriteString(strings.Repeat("─", leftWidth))
	b.WriteString("┬")
	b.WriteString(strings.Repeat("─", rightWidth))
	b.WriteString("\n")

	// 描画する最大行数♥
	maxListLines := min(len(m.filteredEntries), m.height-6)
	maxPreviewLines := len(m.previewCache)
	maxLines := max(maxListLines, maxPreviewLines) // どちらか長い方♠

	for i := 0; i < maxLines; i++ {
		// 左側: ファイルリスト♧
		if i < len(m.filteredEntries) {
			entry := m.filteredEntries[i]
			cursor := "  "
			if i == m.cursor {
				cursor = "\033[1;33m>\033[0m "
			}

			icon := "📄"
			color := "\033[0m"
			if entry.IsDir {
				icon = "📁"
				color = "\033[1;34m"
			}

			displayPath := entry.Name
			if entry.DirPath != "." {
				displayPath = filepath.Join(entry.DirPath, entry.Name)
			}

			cursorWidth := 2 // "  " or "> " どちらも2文字♥
			iconWidth := 2   // 絵文字は2文字幅♧
			spaceWidth := 1  // アイコンと名前の間

			// 表示幅 = カーソル + アイコン + スペース + パス♠
			visibleLen := cursorWidth + iconWidth + spaceWidth + len(displayPath)

			// 切り詰め処理（変更なし）♥
			if visibleLen > leftWidth-1 {
				overflow := visibleLen - (leftWidth - 4)
				if overflow > 0 && len(displayPath) > overflow {
					displayPath = displayPath[:len(displayPath)-overflow] + "..."
				}
			}

			line := fmt.Sprintf("%s%s %s%s\033[0m", cursor, icon, color, displayPath)

			b.WriteString(line)

			// ★パディング計算を修正★♧
			// 切り詰め後の実際の表示幅を再計算♠
			actualVisible := cursorWidth + iconWidth + spaceWidth + len(displayPath)
			padding := leftWidth - actualVisible
			if padding > 0 {
				b.WriteString(strings.Repeat(" ", padding))
			}
		} else {
			// ファイルリストが終わったら空白♠
			b.WriteString(strings.Repeat(" ", leftWidth))
		}

		b.WriteString("│")

		// 右側: プレビュー（独立したインデックス）♥
		if i < len(m.previewCache) {
			previewLine := m.previewCache[i]

			// 右側の幅に収める♧
			if len(previewLine) > rightWidth-2 {
				previewLine = previewLine[:rightWidth-5] + "..."
			}

			b.WriteString(" " + previewLine)
		}

		b.WriteString("\033[K\n") // 行末クリア追加♠
	}

	// フッター♥
	b.WriteString("\n")
	b.WriteString("\033[2m[Ctrl+N/P]移動 [Enter]選択 [Ctrl+D]終了 [Preview: ON]\033[0m")

	return b.String()
}

// HandleInput は入力処理
func (m *Model) HandleInput(r rune) (bool, string, error) {
	switch {
	case r == m.keymap.Quit:
		return true, "", nil // 終了

	case r == m.keymap.Down:
		if m.cursor < len(m.filteredEntries)-1 {
			m.cursor++
			m.updatePreview()
		}

	case r == m.keymap.Up:
		if m.cursor > 0 {
			m.cursor--
			m.updatePreview()
		}

	case r == m.keymap.Enter:
		if len(m.filteredEntries) > 0 {
			selected := m.filteredEntries[m.cursor]
			if selected.IsDir {
				// ディレクトリドリルダウン
				return false, "", m.changeDirectory(selected.Path)
			}
			// ファイル選択: パスを返す
			fullPath := filepath.Join(m.currentDir, selected.Path)
			return true, fullPath, nil
		}

	case r == m.keymap.Backspace || r == m.keymap.DeleteQuery:
		if len(m.query) > 0 {
			m.query = m.query[:len(m.query)-1]
			m.updateFilter()
		}

	default:
		// 通常文字: クエリに追加
		if r >= 32 && r < 127 {
			m.query += string(r)
			m.updateFilter()
		}
	}

	return false, "", nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// getTerminalSize はターミナルサイズを取得
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
		return 80, 24 // デフォルト値
	}
	return int(ws.Col), int(ws.Row)
}
