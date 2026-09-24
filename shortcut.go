package main

import (
	"bufio"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf16"
)

// Максимум, который читаем из ярлыка. Обычный .lnk весит 1–4 КБ,
// поэтому чтение почти мгновенное даже на медленном HDD.
const maxShortcutRead = 64 * 1024

// ShortcutText возвращает текст из ярлыка (.lnk / .url): путь назначения,
// аргументы, URL, путь к иконке. Файл открывается ТОЛЬКО для чтения,
// и программа, на которую указывает ярлык, никак не затрагивается.
func ShortcutText(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()

	switch ext {
	case ".url", ".website":
		return readURLFile(f)
	case ".lnk", ".pif":
		buf, err := io.ReadAll(io.LimitReader(f, maxShortcutRead))
		if err != nil {
			return ""
		}
		return extractStrings(buf)
	}
	return ""
}

// readURLFile читает строки URL= и IconFile= из интернет-ярлыка.
func readURLFile(r io.Reader) string {
	var sb strings.Builder
	sc := bufio.NewScanner(io.LimitReader(r, maxShortcutRead))
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		l := strings.ToLower(line)
		if strings.HasPrefix(l, "url=") || strings.HasPrefix(l, "iconfile=") {
			sb.WriteString(line[strings.IndexByte(line, '=')+1:])
			sb.WriteByte(' ')
		}
	}
	return sb.String()
}

// extractStrings вытаскивает из двоичного .lnk все читаемые строки
// (ASCII и UTF-16LE). Этого достаточно, чтобы узнать путь к программе,
// не используя тяжёлый COM/Shell API.
func extractStrings(b []byte) string {
	var sb strings.Builder
	const minLen = 4

	// ASCII / ANSI строки
	start := -1
	for i := 0; i <= len(b); i++ {
		if i < len(b) && b[i] >= 0x20 && b[i] < 0x7f {
			if start < 0 {
				start = i
			}
			continue
		}
		if start >= 0 && i-start >= minLen {
			sb.Write(b[start:i])
			sb.WriteByte(' ')
		}
		start = -1
	}

	// UTF-16LE строки (включая кириллицу)
	for off := 0; off < 2; off++ {
		var run []uint16
		flush := func() {
			if len(run) >= minLen {
				sb.WriteString(string(utf16.Decode(run)))
				sb.WriteByte(' ')
			}
			run = run[:0]
		}
		for i := off; i+1 < len(b); i += 2 {
			c := uint16(b[i]) | uint16(b[i+1])<<8
			printable := (c >= 0x20 && c < 0x7f) || (c >= 0x0400 && c <= 0x04ff)
			if printable {
				run = append(run, c)
			} else {
				flush()
			}
		}
		flush()
	}
	return sb.String()
}
