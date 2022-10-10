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

type Conf struct {
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
				return "", fmt.Errorf("error 714af37 (%w)", err)
			}
			userConfigDir = path.Join(homeDir, ".config")
		}
	default:
		userConfigDir, err = os.UserConfigDir()
		if err != nil {
			return "", fmt.Errorf("error a12c224 (%w)", err)
		}
	}
	return filepath.Join(userConfigDir, "querysan.toml"), nil
}

//go:embed config-default.toml
var defaultConfigToml []byte

func EnsureConfigFile() error {
	confFilePath, err := configFilePath()
	if err != nil {
		return fmt.Errorf("error 7f87f13 (%w)", err)
	}
	_, err = os.Stat(confFilePath)
	if err == nil {
		log.Printf("%s exists.", confFilePath)
		return err
	}
	confDirPath := path.Dir(confFilePath)
	if err = os.MkdirAll(confDirPath, 0755); err != nil {
		return fmt.Errorf("error 5c4c729 (%w)", err)
	}
	confFile, err := os.OpenFile(confFilePath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("error ff7a84d (%w)", err)
	}
	defer func() { _ = confFile.Close() }()
	reader := bytes.NewReader(defaultConfigToml)
	_, err = io.Copy(confFile, reader)
	if err != nil {
		return fmt.Errorf("error e102197 (%w)", err)
	}
	return nil
}

const homeVariable = "$HOME"

// todo: service provider 式か、context に持たせるかする
var conf *Conf

func ReadConfig() (*Conf, error) {
	configFilePath, err := configFilePath()
	data, err := os.ReadFile(configFilePath)
	if err != nil {
		return nil, fmt.Errorf("error e8afdd7 (%w)", err)
	}
	config := &Conf{}
	err = toml.Unmarshal(data, config)
	if err != nil {
		return nil, fmt.Errorf("error 0453b04 (%w)", err)
	}
	homeVariableWithSeparator := fmt.Sprintf("%v%c", homeVariable, filepath.Separator)
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("error ffbb128 (%w)", err)
	}
	for i, documentDirectory := range config.DocumentDirectories {
		if strings.Index(documentDirectory.Path, homeVariableWithSeparator) == 0 {
			p := strings.Replace(documentDirectory.Path, homeVariable, homeDir, 1)
			p, err := filepath.EvalSymlinks(p)
			if err != nil {
				return nil, fmt.Errorf("error 063503d (%w)", err)
			}
			config.DocumentDirectories[i].Path = p
		}
	}
	conf = config
	return config, nil
}
