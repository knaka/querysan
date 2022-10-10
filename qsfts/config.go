package qsfts

import (
	"bytes"
	_ "embed"
	"fmt"
	"github.com/pelletier/go-toml/v2"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
)

type conf struct {
	DocumentDirectories []*confDocumentDirectory `toml:"document_directories"`
}

type confDocumentDirectory struct {
	Path       string   `toml:"path"`
	Extensions []string `toml:"extensions"`
}

func configFilePath() (string, error) {
	var userConfigDir string
	var err error
	switch runtime.GOOS {
	case "darwin":
		userConfigDir = os.Getenv("XDG_CONFIG_HOME")
		if userConfigDir == "" {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return "", fmt.Errorf("error 698c60c (%w)", err)
			}
			userConfigDir = path.Join(homeDir, ".config")
		}
	default:
		userConfigDir, err = os.UserConfigDir()
		if err != nil {
			return "", fmt.Errorf("error 09b3cbe (%w)", err)
		}
	}
	return filepath.Join(userConfigDir, "querysan.toml"), nil
}

//go:embed config-default.toml
var defaultConfigToml []byte

func EnsureConfigFile() error {
	confFilePath, err := configFilePath()
	if err != nil {
		return fmt.Errorf("error c4fdc44 (%w)", err)
	}
	_, err = os.Stat(confFilePath)
	if err == nil {
		log.Printf("%s exists.", confFilePath)
		return err
	}
	confDirPath := path.Dir(confFilePath)
	if err = os.MkdirAll(confDirPath, 0755); err != nil {
		return fmt.Errorf("error 243aa64 (%w)", err)
	}
	confFile, err := os.OpenFile(confFilePath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("error 778b59e (%w)", err)
	}
	defer func() { _ = confFile.Close() }()
	reader := bytes.NewReader(defaultConfigToml)
	_, err = io.Copy(confFile, reader)
	if err != nil {
		return fmt.Errorf("error fe376dc (%w)", err)
	}
	return nil
}

const homeVariable = "$HOME"

func ReadConfig() (*conf, error) {
	configFilePath, err := configFilePath()
	data, err := os.ReadFile(configFilePath)
	if err != nil {
		return nil, fmt.Errorf("error ee00fa6 (%w)", err)
	}
	config := &conf{}
	err = toml.Unmarshal(data, config)
	if err != nil {
		return nil, fmt.Errorf("error e64a370 (%w)", err)
	}
	homeVariableWithSeparator := fmt.Sprintf("%v%c", homeVariable, filepath.Separator)
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("error 1d4ccd3 (%w)", err)
	}
	for i, documentDirectory := range config.DocumentDirectories {
		if strings.Index(documentDirectory.Path, homeVariableWithSeparator) == 0 {
			p := strings.Replace(documentDirectory.Path, homeVariable, homeDir, 1)
			p, err := filepath.EvalSymlinks(p)
			if err != nil {
				return nil, fmt.Errorf("error a0d42cf (%w)", err)
			}
			config.DocumentDirectories[i].Path = p
		}
	}
	return config, nil
}
