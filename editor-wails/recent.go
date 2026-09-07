package main

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func (a *App) GetRecentProjects() ([]string, error) {
	configDir, _ := os.UserConfigDir()
	data, err := os.ReadFile(filepath.Join(configDir, "janusz-editor", "recent.json"))
	if err!= nil { return []string{}, nil }
	var recent []string
	json.Unmarshal(data, &recent)
	return recent, nil
}

func (a *App) AddRecentProject(path string) error {
	recent, _ := a.GetRecentProjects()
	for i, p := range recent {
		if p == path { recent = append(recent[:i], recent[i+1:]...); break }
	}
	recent = append([]string{path}, recent...)
	if len(recent) > 10 { recent = recent[:10] }
	configDir, _ := os.UserConfigDir()
	configPath := filepath.Join(configDir, "janusz-editor")
	os.MkdirAll(configPath, 0755)
	data, _ := json.MarshalIndent(recent, "", " ")
	runtime.LogInfo(a.ctx, "[APP] Zapisuję recent: "+path)
	return os.WriteFile(filepath.Join(configPath, "recent.json"), data, 0644)
}