package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Имя скрытого файла журнала, который лежит на рабочем столе.
// По нему можно отменить сортировку.
const undoFileName = ".desktop-sorter-undo.json"

type Move struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type Run struct {
	Time    string   `json:"time"`
	Moves   []Move   `json:"moves"`
	Created []string `json:"created_dirs"`
}

type UndoLog struct {
	Runs []Run `json:"runs"`
}

type Planned struct {
	Name     string
	Category string
}

// ignoredNames — системные файлы рабочего стола, которые никогда не трогаем.
var ignoredNames = map[string]bool{
	"desktop.ini": true,
	"thumbs.db":   true,
	undoFileName:  true,
}

// Plan строит список «файл → категория», ничего не перемещая.
// Берутся только файлы, лежащие прямо на рабочем столе.
// Папки, скрытые и системные файлы пропускаются.
func Plan(desktop, selfPath string) ([]Planned, error) {
	entries, err := os.ReadDir(desktop)
	if err != nil {
		return nil, err
	}
	selfPath = strings.ToLower(filepath.Clean(selfPath))
	profile := strings.ToLower(os.Getenv("USERPROFILE"))

	var out []Planned
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || ignoredNames[strings.ToLower(name)] || strings.HasPrefix(name, "~$") {
			continue
		}
		full := filepath.Join(desktop, name)
		if strings.ToLower(filepath.Clean(full)) == selfPath {
			continue // саму программу не перемещаем
		}
		if !e.Type().IsRegular() && e.Type()&os.ModeSymlink == 0 {
			continue
		}
		if isHiddenOrSystem(full) {
			continue
		}
		target := ShortcutText(full)
		if profile != "" {
			// чтобы имя пользователя (например «Gamer») не влияло на категорию
			target = strings.ReplaceAll(strings.ToLower(target), profile, "%userprofile%")
		}
		out = append(out, Planned{Name: name, Category: Classify(name, target)})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Category != out[j].Category {
			return out[i].Category < out[j].Category
		}
		return strings.ToLower(out[i].Name) < strings.ToLower(out[j].Name)
	})
	return out, nil
}

// uniquePath подбирает свободное имя: «файл (2).lnk», «файл (3).lnk»...
func uniquePath(p string) string {
	if _, err := os.Lstat(p); errors.Is(err, os.ErrNotExist) {
		return p
	}
	ext := filepath.Ext(p)
	base := strings.TrimSuffix(p, ext)
	for i := 2; ; i++ {
		c := fmt.Sprintf("%s (%d)%s", base, i, ext)
		if _, err := os.Lstat(c); errors.Is(err, os.ErrNotExist) {
			return c
		}
	}
}

// Apply перемещает файлы по папкам. Используется os.Rename в пределах
// рабочего стола: данные не копируются и не изменяются, меняется только
// расположение ярлыка/файла. Цели ярлыков (программы на C:\ и т.д.) не трогаются.
func Apply(desktop string, plan []Planned, logf func(string, ...interface{})) (int, int) {
	run := Run{Time: time.Now().Format("2006-01-02 15:04:05")}
	moved, failed := 0, 0
	for _, p := range plan {
		dir := filepath.Join(desktop, p.Category)
		if _, err := os.Stat(dir); errors.Is(err, os.ErrNotExist) {
			if err := os.Mkdir(dir, 0o755); err != nil {
				logf("  [ошибка] не удалось создать папку %s: %v\n", p.Category, err)
				failed++
				continue
			}
			run.Created = append(run.Created, dir)
		}
		from := filepath.Join(desktop, p.Name)
		to := uniquePath(filepath.Join(dir, p.Name))
		if err := os.Rename(from, to); err != nil {
			logf("  [пропущен] %s: %v\n", p.Name, err)
			failed++
			continue
		}
		run.Moves = append(run.Moves, Move{From: from, To: to})
		moved++
	}
	if moved > 0 || len(run.Created) > 0 {
		lg := loadUndo(desktop)
		lg.Runs = append(lg.Runs, run)
		if len(lg.Runs) > 20 { // храним только последние 20 запусков
			lg.Runs = lg.Runs[len(lg.Runs)-20:]
		}
		if err := saveUndo(desktop, lg); err != nil {
			logf("  [внимание] не удалось сохранить журнал отмены: %v\n", err)
		}
	}
	return moved, failed
}

// Undo возвращает файлы последнего запуска на рабочий стол.
func Undo(desktop string, logf func(string, ...interface{})) (int, error) {
	lg := loadUndo(desktop)
	if len(lg.Runs) == 0 {
		return 0, errors.New("нечего отменять — журнал пуст")
	}
	run := lg.Runs[len(lg.Runs)-1]
	restored := 0
	for i := len(run.Moves) - 1; i >= 0; i-- {
		m := run.Moves[i]
		// защита: работаем только с путями внутри рабочего стола
		if !inside(desktop, m.From) || !inside(desktop, m.To) {
			continue
		}
		if _, err := os.Lstat(m.To); err != nil {
			logf("  [нет файла] %s\n", filepath.Base(m.To))
			continue
		}
		dst := uniquePath(m.From)
		if err := os.Rename(m.To, dst); err != nil {
			logf("  [ошибка] %s: %v\n", filepath.Base(m.To), err)
			continue
		}
		restored++
	}
	// удаляем только те папки, которые создала программа, и только если они пусты
	for _, d := range run.Created {
		if inside(desktop, d) {
			os.Remove(d) // os.Remove не удалит непустую папку
		}
	}
	lg.Runs = lg.Runs[:len(lg.Runs)-1]
	if len(lg.Runs) == 0 {
		removeUndo(desktop)
	} else {
		saveUndo(desktop, lg)
	}
	return restored, nil
}

func inside(root, p string) bool {
	rel, err := filepath.Rel(root, p)
	if err != nil {
		return false
	}
	return rel != "." && !strings.HasPrefix(rel, "..") && !filepath.IsAbs(rel)
}

func undoPath(desktop string) string { return filepath.Join(desktop, undoFileName) }

func loadUndo(desktop string) UndoLog {
	var lg UndoLog
	b, err := os.ReadFile(undoPath(desktop))
	if err == nil {
		json.Unmarshal(b, &lg)
	}
	return lg
}

func saveUndo(desktop string, lg UndoLog) error {
	p := undoPath(desktop)
	unhide(p) // в Windows скрытый файл нельзя перезаписать напрямую
	b, _ := json.MarshalIndent(lg, "", "  ")
	if err := os.WriteFile(p, b, 0o644); err != nil {
		return err
	}
	hide(p)
	return nil
}

func removeUndo(desktop string) {
	p := undoPath(desktop)
	unhide(p)
	os.Remove(p)
}
