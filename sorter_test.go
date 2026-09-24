package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestClassify(t *testing.T) {
	cases := []struct{ name, target, want string }{
		{"Google Chrome.lnk", `C:\Program Files\Google\Chrome\Application\chrome.exe`, CatBrowsers},
		{"Yandex.lnk", `%userprofile%\appdata\local\yandex\yandexbrowser\application\browser.exe`, CatBrowsers},
		{"Opera GX.lnk", "", CatBrowsers},
		{"Counter-Strike 2.url", "URL=steam://rungameid/730", CatGames},
		{"Мой проект.lnk", `D:\Games\Witcher3\bin\witcher3.exe`, CatGames},
		{"Steam.lnk", `C:\Program Files (x86)\Steam\steam.exe`, CatGames},
		{"Telegram.lnk", `C:\Users\u\AppData\Roaming\Telegram Desktop\Telegram.exe`, CatMessengers},
		{"Discord.lnk", "", CatMessengers},
		{"Word.lnk", `C:\Program Files\Microsoft Office\root\Office16\WINWORD.EXE`, CatOffice},
		{"Visual Studio Code.lnk", "", CatDev},
		{"VLC media player.lnk", "", CatMedia},
		{"WinRAR.lnk", "", CatUtils},
		{"Что-то.lnk", `C:\Program Files\Unknown\app.exe`, CatApps},
		{"ChromeSetup.exe", "", CatInstallers},
		{"отчёт.docx", "", CatDocs},
		{"фото.JPG", "", CatImages},
		{"фильм.mkv", "", CatVideo},
		{"песня.mp3", "", CatAudio},
		{"backup.zip", "", CatArchives},
		{"YouTube.url", "URL=https://youtube.com", CatWeb},
		{"file.xyz", "", CatOther},
	}
	for _, c := range cases {
		if got := Classify(c.name, c.target); got != c.want {
			t.Errorf("%s: got %q want %q", c.name, got, c.want)
		}
	}
}

func TestSortAndUndo(t *testing.T) {
	d := t.TempDir()
	files := map[string]string{
		"Google Chrome.lnk": "C:\\Program Files\\Google\\Chrome\\Application\\chrome.exe",
		"CS2.url":           "[InternetShortcut]\r\nURL=steam://rungameid/730\r\n",
		"notes.txt":         "hello",
		"desktop.ini":       "x",
		"a.jpg":             "img",
	}
	for n, c := range files {
		os.WriteFile(filepath.Join(d, n), []byte(c), 0o644)
	}
	os.Mkdir(filepath.Join(d, "Моя папка"), 0o755)
	os.WriteFile(filepath.Join(d, "Моя папка", "inner.txt"), []byte("x"), 0o644)
	os.Mkdir(filepath.Join(d, CatDocs), 0o755)
	os.WriteFile(filepath.Join(d, CatDocs, "notes.txt"), []byte("old"), 0o644)

	plan, err := Plan(d, filepath.Join(d, "sorter.exe"))
	if err != nil || len(plan) != 4 {
		t.Fatalf("plan: %v %v", plan, err)
	}
	moved, failed := Apply(d, plan, func(string, ...interface{}) {})
	if moved != 4 || failed != 0 {
		t.Fatalf("moved %d failed %d", moved, failed)
	}
	for _, p := range []string{CatBrowsers + "/Google Chrome.lnk", CatGames + "/CS2.url", CatDocs + "/notes (2).txt", CatImages + "/a.jpg", "desktop.ini", "Моя папка/inner.txt", CatDocs + "/notes.txt"} {
		if _, err := os.Stat(filepath.Join(d, p)); err != nil {
			t.Errorf("missing %s", p)
		}
	}
	n, err := Undo(d, func(string, ...interface{}) {})
	if err != nil || n != 4 {
		t.Fatalf("undo %d %v", n, err)
	}
	for _, p := range []string{"Google Chrome.lnk", "CS2.url", "notes.txt", "a.jpg", CatDocs + "/notes.txt"} {
		if _, err := os.Stat(filepath.Join(d, p)); err != nil {
			t.Errorf("after undo missing %s", p)
		}
	}
	for _, p := range []string{CatBrowsers, CatGames, CatImages, undoFileName} {
		if _, err := os.Stat(filepath.Join(d, p)); err == nil {
			t.Errorf("should be removed: %s", p)
		}
	}
	b, _ := os.ReadFile(filepath.Join(d, CatDocs, "notes.txt"))
	if string(b) != "old" {
		t.Error("existing file changed")
	}
}

func TestExtractStrings(t *testing.T) {
	var b []byte
	b = append(b, 0x4c, 0, 0, 0, 1, 2, 3)
	for _, r := range `D:\Игры\Game.exe` {
		b = append(b, byte(r), byte(r>>8))
	}
	b = append(b, 0, 0, 'C', ':', '\\', 'X', '.', 'e', 'x', 'e', 0)
	s := extractStrings(b)
	if !contains(s, `D:\Игры\Game.exe`) || !contains(s, `C:\X.exe`) {
		t.Fatalf("got %q", s)
	}
}

func contains(s, sub string) bool { return strings.Contains(s, sub) }
