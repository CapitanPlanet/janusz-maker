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

// --- Helpers - SECURITY ---

func (a *App) cleanProjectPath() string {
	if a.projectPath == "" {
		return ""
	}
	return filepath.Clean(a.projectPath)
}

func (a *App) isInsideProject(targetPath string) bool {
	if a.projectPath == "" {
		return false
	}
	proj := a.cleanProjectPath()
	target := filepath.Clean(targetPath)
	// must be inside proj or equal
	rel, err := filepath.Rel(proj, target)
	if err != nil {
		return false
	}
	return !strings.HasPrefix(rel, ".."+string(os.PathSeparator)) && rel != ".."
}

func (a *App) safeJoin(elem ...string) (string, error) {
	if a.projectPath == "" {
		return "", fmt.Errorf("brak projektu")
	}
	joined := filepath.Join(elem...)
	clean := filepath.Clean(joined)
	if !a.isInsideProject(clean) && clean != a.cleanProjectPath() {
		return "", fmt.Errorf("path traversal blocked: %s", clean)
	}
	return clean, nil
}

func (a *App) cleanAssetPath(input string) string {
	p := strings.TrimSpace(input)
	if p == "" {
		return ""
	}
	p = strings.ReplaceAll(p, "\\", "/")
	p = filepath.ToSlash(p)
	// If absolute or contains project path, strip to filename
	if filepath.IsAbs(p) || strings.Contains(p, ":") {
		cleanProject := filepath.ToSlash(filepath.Clean(a.projectPath))
		if cleanProject != "" && strings.Contains(p, cleanProject) {
			p = strings.Replace(p, cleanProject, "", 1)
			p = strings.TrimLeft(p, "/")
		} else {
			p = "images/" + filepath.Base(p)
		}
	}
	p = strings.TrimLeft(p, "/")
	return filepath.FromSlash(p)
}

// --- Project ---

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
	path = filepath.Clean(path)
	var cmd *exec.Cmd
	switch goruntime.GOOS {
	case "windows":
		cmd = exec.Command("explorer", path)
	case "darwin":
		cmd = exec.Command("open", path)
	default:
		cmd = exec.Command("xdg-open", path)
	}
	return cmd.Start()
}

func (a *App) CreateProject(basePath string, name string) error {
	basePath = filepath.Clean(basePath)
	// Structure for janusz-maker
	for _, d := range []string{
		"Data",
		"images",
		filepath.Join("sounds", "sfx"),
		filepath.Join("sounds", "voice"),
		filepath.Join("sounds", "music"),
	} {
		if err := os.MkdirAll(filepath.Join(basePath, d), 0755); err != nil {
			return err
		}
	}

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
	if err := os.WriteFile(janprojPath, []byte(janprojContent), 0644); err != nil {
		return err
	}

	// Minimal day1
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
	if err := os.WriteFile(day1Path, []byte(day1Content), 0644); err != nil {
		return err
	}

	// Copy template assets if exist - from embed first, then fallback to disk
	// We try to copy bg_tutorial.webp from frontend/src/assets if available
	srcAssets := "frontend/src/assets"
	if _, err := os.Stat(srcAssets); err == nil {
		_ = filepath.Walk(srcAssets, func(p string, info fs.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return nil
			}
			low := strings.ToLower(info.Name())
			if !(strings.HasSuffix(low, ".webp") || strings.HasSuffix(low, ".png") || strings.HasSuffix(low, ".jpg") || strings.HasSuffix(low, ".jpeg")) {
				return nil
			}
			dst := filepath.Join(basePath, "images", info.Name())
			if _, err := os.Stat(dst); os.IsNotExist(err) {
				if b, err := os.ReadFile(p); err == nil {
					_ = os.WriteFile(dst, b, 0644)
				}
			}
			return nil
		})
	}

	a.projectPath = basePath
	return nil
}

// --- Generic IO ---

func (a *App) ReadJSON(fullPath string) (string, error) {
	b, err := os.ReadFile(filepath.Clean(fullPath))
	return string(b), err
}

func (a *App) WriteJSON(fullPath string, content string) error {
	clean := filepath.Clean(fullPath)
	if err := os.MkdirAll(filepath.Dir(clean), 0755); err != nil {
		return err
	}
	return os.WriteFile(clean, []byte(content), 0644)
}

func (a *App) SaveJsonFile(filename string, content string) error {
	if a.projectPath == "" {
		return fmt.Errorf("brak projectPath")
	}
	base := filepath.Base(filename)
	target := filepath.Join(a.projectPath, "Data", base)
	safe, err := a.safeJoin(target)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(safe), 0755); err != nil {
		return err
	}
	return os.WriteFile(safe, []byte(content), 0644)
}

func (a *App) ListFiles(dirPath string, ext string) ([]string, error) {
	clean := filepath.Clean(dirPath)
	entries, err := os.ReadDir(clean)
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
	if projectPath == "" {
		return fmt.Errorf("brak projektu")
	}
	target := filepath.Join(projectPath, a.cleanAssetPath(relativePath))
	if !a.isInsideProject(filepath.Clean(target)) {
		return fmt.Errorf("delete blocked - outside project")
	}
	return os.Remove(target)
}

// --- Images ---

func (a *App) SelectImageFile() (string, error) {
	return runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Wybierz obraz",
		Filters: []runtime.FileFilter{{DisplayName: "Images", Pattern: "*.png;*.jpg;*.jpeg;*.webp"}},
	})
}

func (a *App) ImportAsset(srcPath string, assetType string) (string, error) {
	if a.projectPath == "" {
		return "", fmt.Errorf("brak projektu")
	}
	if strings.TrimSpace(assetType) == "" {
		assetType = "bg"
	}
	assetType = strings.ToLower(strings.TrimSpace(assetType))

	dstDir := filepath.Join(a.projectPath, "images")
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return "", err
	}

	origBase := filepath.Base(srcPath)
	base := origBase
	lowBase := strings.ToLower(base)

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
	// avoid overwrite - add suffix
	if _, err := os.Stat(dst); err == nil {
		ext := filepath.Ext(base)
		name := strings.TrimSuffix(base, ext)
		for i := 1; i < 100; i++ {
			candidate := filepath.Join(dstDir, fmt.Sprintf("%s_%d%s", name, i, ext))
			if _, err := os.Stat(candidate); os.IsNotExist(err) {
				dst = candidate
				base = filepath.Base(candidate)
				break
			}
		}
	}

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
	if _, err = io.Copy(out, in); err != nil {
		return "", err
	}
	return "images/" + base, nil
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
	_ = filepath.WalkDir(filepath.Join(a.projectPath, "images"), func(p string, d fs.DirEntry, err error) error {
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
	switch ext {
	case ".jpg", ".jpeg":
		mime = "image/jpeg"
	case ".webp":
		mime = "image/webp"
	}
	return fmt.Sprintf("data:%s;base64,%s", mime, base64.StdEncoding.EncodeToString(b)), nil
}

func (a *App) DeleteAsset(relPath string) error {
	if a.projectPath == "" {
		return fmt.Errorf("brak projektu")
	}
	relPath = a.cleanAssetPath(relPath)
	target := filepath.Join(a.projectPath, relPath)
	clean := filepath.Clean(target)
	if !a.isInsideProject(clean) {
		return fmt.Errorf("delete blocked")
	}
	return os.Remove(clean)
}

func (a *App) ListAssets(projectPath string) ([]string, error) {
	if projectPath == "" {
		projectPath = a.projectPath
	}
	if projectPath == "" {
		return []string{}, nil
	}
	var out []string
	imagesDir := filepath.Join(projectPath, "images")
	_ = filepath.WalkDir(imagesDir, func(p string, d fs.DirEntry, err error) error {
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

// --- Audio ---

func (a *App) SelectAudioFile() (string, error) {
	return runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Wybierz audio",
		Filters: []runtime.FileFilter{{DisplayName: "Audio", Pattern: "*.mp3;*.wav;*.ogg"}},
	})
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
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return "", err
	}
	dst := filepath.Join(dstDir, filepath.Base(srcPath))

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
	if _, err := io.Copy(out, in); err != nil {
		return "", err
	}
	return filepath.ToSlash(filepath.Join("sounds", sub, filepath.Base(srcPath))), nil
}

func (a *App) ListAudioAssets(projectPath string) ([]string, error) {
	if projectPath == "" {
		projectPath = a.projectPath
	}
	if projectPath == "" {
		return []string{}, nil
	}
	var out []string
	baseSounds := filepath.Join(projectPath, "sounds")
	_ = filepath.WalkDir(baseSounds, func(p string, d fs.DirEntry, err error) error {
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
	target := filepath.Join(a.projectPath, relPath)
	clean := filepath.Clean(target)
	if !a.isInsideProject(clean) {
		return fmt.Errorf("delete blocked")
	}
	return os.Remove(clean)
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
	switch ext {
	case ".wav":
		mime = "audio/wav"
	case ".ogg":
		mime = "audio/ogg"
	}
	return fmt.Sprintf("data:%s;base64,%s", mime, base64.StdEncoding.EncodeToString(b)), nil
}
