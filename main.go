package main

import (
	"os"
	"path/filepath"

	"github.com/3stadt/vivaldi-url-profile-router/cmd"
)

func main() {
	// Set working directory same as exe location
	exePath, err := os.Executable()
	if err == nil {
		exeDir := filepath.Dir(exePath)
		_ = os.Chdir(exeDir)
	}
	// call cobra
	cmd.Execute()
}
