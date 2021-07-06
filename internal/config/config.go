package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

func ParseFile(out interface{}, path string) error {
	p, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	f, err := os.Open(p)
	if err != nil {
		return fmt.Errorf("failed to load YAML config file: %w", err)
	}
	defer f.Close()
	b, err := ioutil.ReadAll(f)
	if err != nil {
		return fmt.Errorf("failed to load YAML config file: %w", err)
	}
	return Parse(out, b)
}

func Parse(out interface{}, config []byte) error {
	err := json.Unmarshal(config, out)
	if err != nil {
		return err
	}
	return nil
}
