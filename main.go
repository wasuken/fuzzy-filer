package main

import (
	"bufio"
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

func main() {
	// 起動ディレクトリ取得
	startDir := "."
	if len(os.Args) > 1 {
		startDir = os.Args[1]
	}

	// モデル初期化
	model, err := NewModel(startDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// /dev/ttyを開く（パイプライン対応）
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening /dev/tty: %v\n", err)
		os.Exit(1)
	}
	defer tty.Close()

	// ターミナルをraw modeに
	oldState, err := setRawModeForFd(int(tty.Fd()))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error setting raw mode: %v\n", err)
		os.Exit(1)
	}

	// 画面クリア & カーソル非表示
	fmt.Fprint(tty, "\033[2J\033[H\033[?25l")

	// 初期描画
	renderToTTY(model, tty)

	// メインループ
	reader := bufio.NewReader(tty)
	var selectedPath string

	for {
		r, _, err := reader.ReadRune()
		if err != nil {
			break
		}

		// 入力処理
		quit, path, err := model.HandleInput(r)
		if err != nil {
			// エラーを画面に表示して継続
			fmt.Fprintf(tty, "\n\033[1;31mError: %v\033[0m\n", err)
			continue
		}

		if quit {
			selectedPath = path
			break
		}

		// 再描画
		renderToTTY(model, tty)
	}

	// 終了時クリーンアップ
	// 1. 画面クリア
	fmt.Fprint(tty, "\033[2J\033[H")

	// 2. ターミナル状態を復元
	restoreTerminalForFd(int(tty.Fd()), oldState)

	// 3. カーソル表示
	fmt.Fprint(tty, "\033[?25h")

	// 4. パスを標準出力に出力（ttyではなくstdout）
	if selectedPath != "" {
		fmt.Println(selectedPath)
	}
}

// renderToTTY は画面を再描画（tty出力版）
func renderToTTY(m *Model, tty *os.File) {
	// カーソルを左上に移動して描画
	fmt.Fprint(tty, "\033[H")
	fmt.Fprint(tty, m.View())
	fmt.Fprint(tty, "\033[K") // 行末までクリア
}

// setRawModeForFd は指定fdをrawモードに設定
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

// restoreTerminalForFd はターミナル状態を復元
func restoreTerminalForFd(fd int, oldState *syscall.Termios) {
	syscall.Syscall6(syscall.SYS_IOCTL, uintptr(fd),
		uintptr(syscall.TCSETS), uintptr(unsafe.Pointer(oldState)),
		0, 0, 0)
}
