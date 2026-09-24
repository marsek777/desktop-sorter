package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
)

const version = "1.0.0"

func printf(format string, a ...interface{}) { writeOut(fmt.Sprintf(format, a...)) }

var stdin = bufio.NewReader(os.Stdin)

func readLine() string {
	s, _ := stdin.ReadString('\n')
	return strings.TrimSpace(s)
}

func pause() {
	printf("\nНажмите Enter, чтобы продолжить...")
	readLine()
}

func showPlan(plan []Planned) {
	if len(plan) == 0 {
		printf("\nНа рабочем столе нет файлов для сортировки.\n")
		return
	}
	cur := ""
	for _, p := range plan {
		if p.Category != cur {
			cur = p.Category
			printf("\n[%s]\n", cur)
		}
		printf("   %s\n", p.Name)
	}
	printf("\nВсего файлов: %d\n", len(plan))
}

func doSort(desktop, self string, ask bool) {
	plan, err := Plan(desktop, self)
	if err != nil {
		printf("Ошибка чтения рабочего стола: %v\n", err)
		return
	}
	showPlan(plan)
	if len(plan) == 0 {
		return
	}
	if ask {
		printf("\nРазложить эти файлы по папкам? (д/н): ")
		a := strings.ToLower(readLine())
		if a != "д" && a != "y" && a != "да" && a != "yes" && a != "l" {
			printf("Отменено.\n")
			return
		}
	}
	moved, failed := Apply(desktop, plan, printf)
	refreshDesktop()
	printf("\nГотово. Перемещено: %d, пропущено: %d.\n", moved, failed)
	printf("Всё можно вернуть пунктом «Отменить последнюю сортировку».\n")
}

func doUndo(desktop string) {
	n, err := Undo(desktop, printf)
	if err != nil {
		printf("%v\n", err)
		return
	}
	refreshDesktop()
	printf("\nВозвращено на рабочий стол файлов: %d.\n", n)
}

func main() {
	initConsole()
	auto := flag.Bool("auto", false, "отсортировать без вопросов")
	undo := flag.Bool("undo", false, "отменить последнюю сортировку")
	preview := flag.Bool("preview", false, "только показать, что куда будет перемещено")
	flag.Parse()

	desktop := desktopPath()
	self, _ := os.Executable()
	if _, err := os.Stat(desktop); err != nil {
		printf("Не найден рабочий стол: %s\n", desktop)
		os.Exit(1)
	}

	switch {
	case *undo:
		doUndo(desktop)
		return
	case *preview:
		plan, _ := Plan(desktop, self)
		showPlan(plan)
		return
	case *auto:
		doSort(desktop, self, false)
		return
	}

	for {
		printf("\n==============================================\n")
		printf("  Desktop Sorter %s — сортировка рабочего стола\n", version)
		printf("==============================================\n")
		printf("Рабочий стол: %s\n\n", desktop)
		printf("  1. Отсортировать рабочий стол\n")
		printf("  2. Предпросмотр (ничего не перемещать)\n")
		printf("  3. Отменить последнюю сортировку\n")
		printf("  0. Выход\n\n")
		printf("Выберите пункт: ")
		switch readLine() {
		case "1":
			doSort(desktop, self, true)
			pause()
		case "2":
			plan, err := Plan(desktop, self)
			if err != nil {
				printf("Ошибка: %v\n", err)
			}
			showPlan(plan)
			pause()
		case "3":
			doUndo(desktop)
			pause()
		case "0":
			return
		}
	}
}
