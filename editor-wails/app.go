package main

import (
	"context"
	"embed"
	"encoding/base64"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	goruntime "runtime"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed frontend/src/assets
//go:embed frontend/templates
var templatesFS embed.FS

type App struct {
	ctx         context.Context
	projectPath string
}

func NewApp() *App { return &App{} }
func (a *App) startup(ctx context.Context) { a.ctx = ctx }
func (a *App) SetProjectPath(path string) { a.projectPath = filepath.Clean(path) }
func (a *App) GetProjectPath() string { return a.projectPath }
func (a *App) GetDefaultProjectPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "Documents", "JanuszProjects")
}
func (a *App) SelectFolder() (string, error) {
	return runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{Title: "Wybierz folder projektu"})
}
func (a *App) OpenProjectFolder(path string) error {
	if path == "" {
		path = a.projectPath
	}
	if path == "" {
		return fmt.Errorf("brak ścieżki")
	}
	var cmd *exec.Cmd
	switch goruntime.GOOS {
	case "windows":
		cmd = exec.Command("explorer", filepath.Clean(path))
	case "darwin":
		cmd = exec.Command("open", path)
	default:
		cmd = exec.Command("xdg-open", path)
	}
	return cmd.Start()
}

func (a *App) CreateProject(basePath string, name string) error {
	basePath = filepath.Clean(basePath)
	os.MkdirAll(filepath.Join(basePath, "Data"), 0755)
	os.MkdirAll(filepath.Join(basePath, "images"), 0755)
	os.MkdirAll(filepath.Join(basePath, "sounds"), 0755)
	os.MkdirAll(filepath.Join(basePath, "sounds", "sfx"), 0755)
	os.MkdirAll(filepath.Join(basePath, "sounds", "voice"), 0755)
	os.MkdirAll(filepath.Join(basePath, "sounds", "music"), 0755)

	janprojPath := filepath.Join(basePath, "project.janproj")
	janprojContent := fmt.Sprintf(`{
  "gameName": "%s",
  "author": "",
  "version": "1.0.0",
  "engineVersion": "2.0.0",
  "startDay": "day1",
  "startScene": "start",
  "statsSystem": {
    "stats": [
      {"id": "CEBULA", "name": "CEBULA", "initial": 0},
      {"id": "WSTYD", "name": "WSTYD", "initial": 0},
      {"id": "PORTFEL", "name": "PORTFEL", "initial": 0},
      {"id": "REPUTACJA", "name": "REPUTACJA", "initial": 0}
    ]
  },
  "avatarSystem": {"default": "", "rules": []}
}`, name)
	os.WriteFile(janprojPath, []byte(janprojContent), 0644)

	day1Path := filepath.Join(basePath, "Data", "day1.json")
	day1Content := `[
  {
    "Id": "start",
    "SceneTitle": "Dzień 1 - Start",
    "Background": "images/bg_tutorial.webp",
    "Text": "Janusz budzi się.",
    "Choices": [{"Text": "Dalej", "Next": "koniec_dnia_1"}],
    "Type": "normal",
    "Day": 1
  },
  {
    "Id": "koniec_dnia_1",
    "SceneTitle": "KONIEC DNIA 1",
    "Background": "images/bg_tutorial.webp",
    "Text": "Koniec dnia 1.",
    "IsEndDay": true,
    "Type": "end_of_day",
    "Day": 1,
    "NextDayId": "day2",
    "NextDay": "day2",
    "Choices": [{"Text": "Śpij", "Next": "END_DAY", "NextDayId": "day2"}]
  }
]`
	os.WriteFile(day1Path, []byte(day1Content), 0644)

	srcAssets := "frontend/src/assets"
	if _, err := os.Stat(srcAssets); err == nil {
		filepath.Walk(srcAssets, func(p string, info fs.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return nil
			}
			low := strings.ToLower(info.Name())
			if !(strings.HasSuffix(low, ".webp") || strings.HasSuffix(low, ".png") || strings.HasSuffix(low, ".jpg") || strings.HasSuffix(low, ".jpeg")) {
				return nil
			}
			dst := filepath.Join(basePath, "images", info.Name())
			if _, err := os.Stat(dst); os.IsNotExist(err) {
				b, _ := os.ReadFile(p)
				os.WriteFile(dst, b, 0644)
			}
			return nil
		})
	}
	a.projectPath = basePath
	return nil
}

func (a *App) ReadJSON(fullPath string) (string, error) {
	b, err := os.ReadFile(filepath.Clean(fullPath))
	return string(b), err
}
func (a *App) WriteJSON(fullPath string, content string) error {
	os.MkdirAll(filepath.Dir(filepath.Clean(fullPath)), 0755)
	return os.WriteFile(filepath.Clean(fullPath), []byte(content), 0644)
}
func (a *App) SaveJsonFile(filename string, content string) error {
	if a.projectPath == "" {
		return fmt.Errorf("brak projectPath")
	}
	os.MkdirAll(filepath.Join(a.projectPath, "Data"), 0755)
	return os.WriteFile(filepath.Join(a.projectPath, "Data", filepath.Base(filename)), []byte(content), 0644)
}
func (a *App) ListFiles(dirPath string, ext string) ([]string, error) {
	entries, err := os.ReadDir(filepath.Clean(dirPath))
	if err != nil {
		return []string{}, nil
	}
	var out []string
	for _, e := range entries {
		if !e.IsDir() && (ext == "" || strings.HasSuffix(strings.ToLower(e.Name()), strings.ToLower(ext))) {
			out = append(out, e.Name())
		}
	}
	return out, nil
}
func (a *App) DeleteFile(projectPath string, relativePath string) error {
	if projectPath == "" {
		projectPath = a.projectPath
	}
	return os.Remove(filepath.Join(projectPath, relativePath))
}
func (a *App) SelectImageFile() (string, error) {
	return runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{Title: "Wybierz obraz", Filters: []runtime.FileFilter{{DisplayName: "Images", Pattern: "*.png;*.jpg;*.jpeg;*.webp"}}})
}

// --- NOWY ImportAsset z typem bg/re/av ---
func (a *App) ImportAsset(srcPath string, assetType string) (string, error) {
	if a.projectPath == "" {
		return "", fmt.Errorf("brak projektu")
	}
	if strings.TrimSpace(assetType) == "" {
		assetType = "bg"
	}
	assetType = strings.ToLower(strings.TrimSpace(assetType))

	dstDir := filepath.Join(a.projectPath, "images")
	os.MkdirAll(dstDir, 0755)

	origBase := filepath.Base(srcPath)
	base := origBase
	lowBase := strings.ToLower(base)

	// jeśli nie ma prefixu, nadaj go
	hasPrefix := strings.HasPrefix(lowBase, "bg_") || strings.HasPrefix(lowBase, "re_") || strings.HasPrefix(lowBase, "av_")
	if !hasPrefix {
		switch assetType {
		case "av", "avatar", "avatars":
			base = "av_" + origBase
		case "re", "reaction", "reakcja":
			base = "re_" + origBase
		default:
			base = "bg_" + origBase
		}
	}

	dst := filepath.Join(dstDir, base)
	in, err := os.Open(srcPath)
	if err != nil {
		return "", err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return "", err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	if err != nil {
		return "", err
	}
	return "images/" + base, nil
}

func (a *App) cleanAssetPath(input string) string {
	p := strings.TrimSpace(input)
	if p == "" {
		return ""
	}
	p = strings.ReplaceAll(p, "\\", "/")
	p = filepath.ToSlash(p)
	if filepath.IsAbs(p) || strings.Contains(p, ":") {
		cleanProject := filepath.ToSlash(filepath.Clean(a.projectPath))
		if strings.Contains(p, cleanProject) {
			p = strings.Replace(p, cleanProject, "", 1)
			p = strings.TrimLeft(p, "/")
		} else {
			p = "images/" + filepath.Base(p)
		}
	}
	p = strings.TrimLeft(p, "/")
	return filepath.FromSlash(p)
}

func (a *App) findFileByName(name string) (string, bool) {
	base := filepath.Base(name)
	if base == "" || base == "." {
		return "", false
	}
	candidate := filepath.Join(a.projectPath, "images", base)
	if _, err := os.Stat(candidate); err == nil {
		return candidate, true
	}
	var found string
	filepath.WalkDir(filepath.Join(a.projectPath, "images"), func(p string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() && strings.EqualFold(d.Name(), base) {
			found = p
			return io.EOF
		}
		return nil
	})
	if found != "" {
		return found, true
	}
	candidate = filepath.Join(a.projectPath, base)
	if _, err := os.Stat(candidate); err == nil {
		return candidate, true
	}
	return "", false
}

func (a *App) GetImageBase64(relPath string) (string, error) {
	if a.projectPath == "" {
		return "", fmt.Errorf("brak projektu")
	}
	if strings.TrimSpace(relPath) == "" {
		return "", fmt.Errorf("pusty asset")
	}
	relPath = a.cleanAssetPath(relPath)
	full := filepath.Join(a.projectPath, relPath)
	full = filepath.Clean(full)
	if _, err := os.Stat(full); os.IsNotExist(err) {
		if found, ok := a.findFileByName(relPath); ok {
			full = found
		}
	}
	b, err := os.ReadFile(full)
	if err != nil {
		return "", fmt.Errorf("nie znaleziono: %s (szukano: %s)", relPath, full)
	}
	ext := strings.ToLower(filepath.Ext(full))
	mime := "image/png"
	if ext == ".jpg" || ext == ".jpeg" {
		mime = "image/jpeg"
	}
	if ext == ".webp" {
		mime = "image/webp"
	}
	return fmt.Sprintf("data:%s;base64,%s", mime, base64.StdEncoding.EncodeToString(b)), nil
}

func (a *App) DeleteAsset(relPath string) error {
	if a.projectPath == "" {
		return fmt.Errorf("brak projektu")
	}
	relPath = a.cleanAssetPath(relPath)
	return os.Remove(filepath.Join(a.projectPath, relPath))
}

func (a *App) ListAssets(projectPath string) ([]string, error) {
	if projectPath == "" {
		projectPath = a.projectPath
	}
	if projectPath == "" {
		return []string{}, nil
	}
	var out []string
	filepath.WalkDir(filepath.Join(projectPath, "images"), func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		low := strings.ToLower(d.Name())
		if strings.HasSuffix(low, ".webp") || strings.HasSuffix(low, ".png") || strings.HasSuffix(low, ".jpg") || strings.HasSuffix(low, ".jpeg") {
			rel, _ := filepath.Rel(projectPath, p)
			out = append(out, filepath.ToSlash(rel))
		}
		return nil
	})
	return out, nil
}

func (a *App) SelectAudioFile() (string, error) {
	return runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{Title: "Wybierz audio", Filters: []runtime.FileFilter{{DisplayName: "Audio", Pattern: "*.mp3;*.wav;*.ogg"}}})
}

func (a *App) ImportAudioAsset(srcPath string, audioType string) (string, error) {
	if a.projectPath == "" {
		return "", fmt.Errorf("brak projektu")
	}
	if strings.TrimSpace(audioType) == "" {
		audioType = "sfx"
	}
	sub := strings.ToLower(strings.TrimSpace(audioType))
	if sub != "sfx" && sub != "voice" && sub != "music" {
		sub = "sfx"
	}
	dstDir := filepath.Join(a.projectPath, "sounds", sub)
	os.MkdirAll(dstDir, 0755)
	dst := filepath.Join(dstDir, filepath.Base(srcPath))
	in, _ := os.Open(srcPath)
	defer in.Close()
	out, _ := os.Create(dst)
	defer out.Close()
	io.Copy(out, in)
	return filepath.ToSlash(filepath.Join("sounds", sub, filepath.Base(srcPath))), nil
}

func (a *App) ListAudioAssets(projectPath string) ([]string, error) {
	if projectPath == "" {
		projectPath = a.projectPath
	}
	var out []string
	baseSounds := filepath.Join(projectPath, "sounds")
	filepath.WalkDir(baseSounds, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		low := strings.ToLower(d.Name())
		if strings.HasSuffix(low, ".mp3") || strings.HasSuffix(low, ".wav") || strings.HasSuffix(low, ".ogg") {
			rel, _ := filepath.Rel(projectPath, p)
			out = append(out, filepath.ToSlash(rel))
		}
		return nil
	})
	return out, nil
}

func (a *App) DeleteAudioAsset(relPath string) error {
	if a.projectPath == "" {
		return fmt.Errorf("brak projektu")
	}
	relPath = a.cleanAssetPath(relPath)
	return os.Remove(filepath.Join(a.projectPath, relPath))
}

func (a *App) GetAudioBase64(relPath string) (string, error) {
	if a.projectPath == "" {
		return "", fmt.Errorf("brak projektu")
	}
	if strings.TrimSpace(relPath) == "" {
		return "", fmt.Errorf("pusty asset")
	}
	relPath = a.cleanAssetPath(relPath)
	full := filepath.Join(a.projectPath, relPath)
	full = filepath.Clean(full)
	if _, err := os.Stat(full); os.IsNotExist(err) {
		if found, ok := a.findFileByName(relPath); ok {
			full = found
		} else {
			full = filepath.Join(a.projectPath, "sounds", filepath.Base(relPath))
		}
	}
	b, err := os.ReadFile(full)
	if err != nil {
		return "", err
	}
	ext := strings.ToLower(filepath.Ext(full))
	mime := "audio/mpeg"
	if ext == ".wav" {
		mime = "audio/wav"
	}
	if ext == ".ogg" {
		mime = "audio/ogg"
	}
	return fmt.Sprintf("data:%s;base64,%s", mime, base64.StdEncoding.EncodeToString(b)), nil
}