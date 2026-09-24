package main

import (
	"path/filepath"
	"strings"
)

// Названия папок-категорий, которые создаются на рабочем столе.
const (
	CatGames      = "Игры"
	CatBrowsers   = "Браузеры"
	CatMessengers = "Мессенджеры"
	CatOffice     = "Офис"
	CatDev        = "Разработка"
	CatGraphics   = "Графика и монтаж"
	CatMedia      = "Музыка и видео (программы)"
	CatUtils      = "Утилиты"
	CatApps       = "Программы"
	CatWeb        = "Ссылки на сайты"
	CatDocs       = "Документы"
	CatImages     = "Изображения"
	CatVideo      = "Видео"
	CatAudio      = "Музыка"
	CatArchives   = "Архивы"
	CatInstallers = "Установщики"
	CatCode       = "Код"
	CatTorrents   = "Торренты"
	CatOther      = "Прочее"
)

// AllCategories — список всех папок, которые программа считает «своими».
// Существующие папки с этими именами не трогаются, в них только добавляются файлы.
var AllCategories = []string{
	CatGames, CatBrowsers, CatMessengers, CatOffice, CatDev, CatGraphics,
	CatMedia, CatUtils, CatApps, CatWeb, CatDocs, CatImages, CatVideo,
	CatAudio, CatArchives, CatInstallers, CatCode, CatTorrents, CatOther,
}

type rule struct {
	cat      string
	keywords []string
}

// Правила для ярлыков и программ. Порядок важен: первое совпадение побеждает.
// Ключевые слова ищутся в имени ярлыка и в пути, на который он указывает.
var appRules = []rule{
	{CatGames, []string{
		"steamapps", "steam://", "\\steam\\", "steam.exe", "steam",
		"epic games", "epicgames", "com.epicgames.launcher", "riot games", "riotclient",
		"battle.net", "battlenet", "blizzard", "gog galaxy", "goggalaxy", "gog.com",
		"origin.exe", "\\origin\\", "ea desktop", "ea app", "electronic arts",
		"ubisoft", "uplay", "rockstar games", "minecraft", "tlauncher", "roblox",
		"wargaming", "world of tanks", "world of warships", "lesta", "мир танков",
		"genshin", "hoyoverse", "mihoyo", "honkai", "valorant", "league of legends",
		"dota", "counter-strike", "cs2", "csgo", "fortnite", "pubg", "apex legends",
		"gta", "grand theft auto", "the sims", "fifa", "ea sports", "warface",
		"vk play", "vkplay", "my.games", "mygames", "game center", "игровой центр",
		"xbox", "\\games\\", "\\игры\\", "game", "игра", "cyberpunk", "witcher",
		"skyrim", "fallout", "terraria", "osu!", "osu.exe", "among us", "hearthstone",
		"overwatch", "diablo", "starcraft", "warcraft", "rust.exe", "\\rust\\", "dayz", "stalker", "сталкер",
	}},
	{CatBrowsers, []string{
		"chrome", "firefox", "mozilla", "opera", "yandexbrowser", "yandex browser",
		"яндекс браузер", "яндекс.браузер", "msedge", "microsoft edge", "brave",
		"vivaldi", "tor browser", "torbrowser", "chromium", "waterfox", "librewolf",
		"palemoon", "pale moon", "maxthon", "atom browser", "arc.exe", "floorp",
		"thorium", "internet explorer", "iexplore", "браузер", "browser",
	}},
	{CatMessengers, []string{
		"telegram", "телеграм", "discord", "whatsapp", "viber", "skype", "zoom",
		"teams", "slack", "signal", "vk messenger", "vk мессенджер", "icq", "element.exe", "\\element\\",
		"thunderbird", "mail.ru agent", "агент mail", "trueconf", "mattermost", "rocket.chat",
	}},
	{CatOffice, []string{
		"winword", "microsoft word", "word 20", "excel", "powerpnt", "powerpoint",
		"onenote", "outlook", "msaccess", "libreoffice", "openoffice", "soffice",
		"wps office", "\\wps", "myoffice", "мойофис", "мой офис", "р7-офис", "r7-office",
		"onlyoffice", "acrobat", "acrord", "foxit", "pdf", "notion", "obsidian",
		"evernote", "abbyy", "finereader", "1cv8", "1с", "1c:предприятие", "office",
		"блокнот",
	}},
	{CatDev, []string{
		"visual studio", "vs code", "vscode", "\\code.exe", "devenv", "pycharm",
		"intellij", "jetbrains", "webstorm", "phpstorm", "clion", "rider", "goland",
		"android studio", "github desktop", "git bash", "\\git\\", "notepad++",
		"sublime", "unity hub", "unity.exe", "\\unity\\", "unreal", "godot", "docker",
		"postman", "cursor", "windsurf", "arduino", "python", "anaconda",
		"eclipse", "netbeans", "codeblocks", "code::blocks", "dev-c++", "putty",
		"winscp", "filezilla", "dbeaver", "pgadmin", "mysql", "xampp", "openserver",
		"virtualbox", "vmware", "wsl", "terminal", "powershell", "cmd.exe",
	}},
	{CatGraphics, []string{
		"photoshop", "illustrator", "lightroom", "premiere", "after effects",
		"afterfx", "indesign", "adobe", "gimp", "krita", "paint.net", "paintdotnet",
		"blender", "figma", "inkscape", "coreldraw", "corel", "davinci", "resolve.exe",
		"capcut", "vegas", "movavi", "filmora", "shotcut", "kdenlive", "handbrake",
		"autocad", "3ds max", "maya", "cinema 4d", "zbrush", "sketchup", "компас",
		"kompas", "paint", "clipchamp", "canva",
	}},
	{CatMedia, []string{
		"vlc", "potplayer", "mpc-hc", "mpc-be", "media player classic", "aimp",
		"spotify", "itunes", "winamp", "foobar", "kmplayer", "gom player",
		"яндекс музыка", "yandex music", "яндекс.музыка", "obs studio", "obs64",
		"\\obs-studio\\", "audacity", "fl studio", "flstudio", "ableton", "reaper",
		"windows media", "wmplayer", "mpv", "kodi", "plex", "bandicam", "fraps",
		"nvidia broadcast", "voicemeeter",
	}},
	{CatUtils, []string{
		"winrar", "7-zip", "7zfm", "bandizip", "peazip", "ccleaner", "qbittorrent",
		"utorrent", "bittorrent", "transmission", "anydesk", "teamviewer",
		"rustdesk", "ammyy", "cpu-z", "cpuz", "gpu-z", "gpuz", "hwmonitor", "hwinfo",
		"msi afterburner", "afterburner", "aida64", "crystaldisk", "everything",
		"total commander", "totalcmd", "far manager", "nvidia", "geforce", "radeon",
		"amd software", "realtek", "kaspersky", "касперск", "avast", "avg", "eset",
		"nod32", "drweb", "dr.web", "malwarebytes", "360 total", "defender",
		"revo uninstaller", "iobit", "driver booster", "driverpack", "rufus",
		"ultraiso", "daemon tools", "etcher", "acronis", "victoria", "recuva",
		"unlocker", "punto switcher", "lightshot", "sharex", "greenshot", "shadowplay",
		"vpn", "outline", "amnezia", "hiddify", "v2ray", "proxifier", "zapret",
		"goodbyedpi", "control panel", "панель управления", "диспетчер",
		"regedit", "taskmgr", "calc", "калькулятор", "snipping", "ножницы",
		"logitech", "razer", "steelseries", "corsair", "hyperx", "bloody",
	}},
}

var extCategories = map[string]string{
	// документы
	".doc": CatDocs, ".docx": CatDocs, ".docm": CatDocs, ".odt": CatDocs, ".rtf": CatDocs,
	".txt": CatDocs, ".md": CatDocs, ".pdf": CatDocs, ".djvu": CatDocs, ".fb2": CatDocs,
	".epub": CatDocs, ".mobi": CatDocs, ".xls": CatDocs, ".xlsx": CatDocs, ".xlsm": CatDocs,
	".ods": CatDocs, ".csv": CatDocs, ".ppt": CatDocs, ".pptx": CatDocs, ".odp": CatDocs,
	".xps": CatDocs, ".one": CatDocs, ".log": CatDocs,
	// изображения
	".jpg": CatImages, ".jpeg": CatImages, ".png": CatImages, ".gif": CatImages,
	".bmp": CatImages, ".webp": CatImages, ".tif": CatImages, ".tiff": CatImages,
	".svg": CatImages, ".ico": CatImages, ".heic": CatImages, ".psd": CatImages,
	".raw": CatImages, ".cr2": CatImages, ".nef": CatImages, ".avif": CatImages,
	// видео
	".mp4": CatVideo, ".mkv": CatVideo, ".avi": CatVideo, ".mov": CatVideo,
	".wmv": CatVideo, ".flv": CatVideo, ".webm": CatVideo, ".m4v": CatVideo,
	".3gp": CatVideo, ".mpg": CatVideo, ".mpeg": CatVideo, ".ts": CatVideo,
	// музыка
	".mp3": CatAudio, ".wav": CatAudio, ".flac": CatAudio, ".ogg": CatAudio,
	".m4a": CatAudio, ".aac": CatAudio, ".wma": CatAudio, ".opus": CatAudio,
	".mid": CatAudio, ".midi": CatAudio,
	// архивы
	".zip": CatArchives, ".rar": CatArchives, ".7z": CatArchives, ".tar": CatArchives,
	".gz": CatArchives, ".bz2": CatArchives, ".xz": CatArchives, ".iso": CatArchives,
	".cab": CatArchives, ".tgz": CatArchives,
	// установщики
	".msi": CatInstallers, ".msix": CatInstallers, ".appx": CatInstallers,
	".appxbundle": CatInstallers, ".msixbundle": CatInstallers, ".apk": CatInstallers,
	// код
	".py": CatCode, ".js": CatCode, ".ts_": CatCode, ".html": CatCode, ".htm": CatCode,
	".css": CatCode, ".cpp": CatCode, ".c": CatCode, ".h": CatCode, ".cs": CatCode,
	".java": CatCode, ".go": CatCode, ".rs": CatCode, ".php": CatCode, ".json": CatCode,
	".xml": CatCode, ".yml": CatCode, ".yaml": CatCode, ".bat": CatCode, ".cmd": CatCode,
	".ps1": CatCode, ".sh": CatCode, ".sql": CatCode, ".ipynb": CatCode, ".lua": CatCode,
	// торренты
	".torrent": CatTorrents,
}

var installerWords = []string{"setup", "install", "installer", "установ", "инсталл", "_x64", "_x86", "-x64", "-x86", "win64", "win32"}

func matchApp(text string) string {
	for _, r := range appRules {
		for _, kw := range r.keywords {
			if strings.Contains(text, kw) {
				return r.cat
			}
		}
	}
	return ""
}

// Classify определяет категорию файла.
// name — имя файла, target — текст, извлечённый из ярлыка (путь, аргументы, URL).
func Classify(name, target string) string {
	lname := strings.ToLower(name)
	ext := filepath.Ext(lname)
	base := strings.TrimSuffix(lname, ext)

	switch ext {
	case ".lnk", ".pif", ".appref-ms":
		// Сначала по имени ярлыка, затем по пути назначения.
		if c := matchApp(base); c != "" {
			return c
		}
		if c := matchApp(strings.ToLower(target)); c != "" {
			return c
		}
		return CatApps
	case ".url", ".website":
		lt := strings.ToLower(target)
		if c := matchApp(base); c != "" {
			return c
		}
		// Игровые ярлыки Steam/Epic/и т.п. имеют вид steam://rungameid/...
		if strings.Contains(lt, "steam://") || strings.Contains(lt, "com.epicgames.launcher://") ||
			strings.Contains(lt, "uplay://") || strings.Contains(lt, "battlenet://") ||
			strings.Contains(lt, "goggalaxy://") || strings.Contains(lt, "origin://") ||
			strings.Contains(lt, "riotclient") || strings.Contains(lt, "steamapps") {
			return CatGames
		}
		if c := matchApp(lt); c != "" && c != CatBrowsers {
			return c
		}
		return CatWeb
	case ".exe":
		for _, w := range installerWords {
			if strings.Contains(base, w) {
				return CatInstallers
			}
		}
		if c := matchApp(base); c != "" {
			return c
		}
		return CatApps
	}
	if c, ok := extCategories[ext]; ok {
		return c
	}
	return CatOther
}
