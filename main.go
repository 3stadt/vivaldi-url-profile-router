package main

import (
	"errors"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/3stadt/vivaldi-url-profile-router/cmd"
)

func main() {
	findConfigDir()
	// call cobra
	cmd.Execute()
}

// Set working directory same as exe location when no config file is present, fail if none found in any possible place
func findConfigDir() {
	userdir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	f, err := os.OpenFile(filepath.Join(userdir, "vupr_error.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("Error opening file: %v", err)
	}
	defer f.Close()
	log.SetOutput(f)

	configDir := "config"
	configFile := "app.yaml"
	workDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("can't get working directory: %s", err)
	}

	if _, err := os.Stat(filepath.Join(workDir, configDir, configFile)); err == nil {
		_ = os.Chdir(workDir)
		return
	}

	_, callerFile, _, ok := runtime.Caller(0)
	if ok {
		callerDir := filepath.Dir(callerFile)
		if _, err := os.Stat(filepath.Join(callerDir, configDir, configFile)); err == nil {
			_ = os.Chdir(callerDir)
			return
		}
	}

	exePath, err := os.Executable()
	if err != nil {
		log.Fatalf("can't get executable directory: %s", err)
	}

	cfgDir := filepath.Dir(exePath)
	_ = os.Chdir(cfgDir)

	if _, err := os.Stat(filepath.Join(cfgDir, configDir, configFile)); errors.Is(err, os.ErrNotExist) {
		log.Fatalf("Can't find %q, make sure it exists in %q", filepath.Join(configDir, configFile), workDir)
	}
}
