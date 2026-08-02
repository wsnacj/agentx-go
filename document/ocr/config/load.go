package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Load 从 YAML 文件加载 ServiceConfig。
func Load(path string) (ServiceConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return ServiceConfig{}, fmt.Errorf("read config %s: %w", path, err)
	}
	var cfg ServiceConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return ServiceConfig{}, fmt.Errorf("unmarshal config %s: %w", path, err)
	}
	if err := cfg.Validate(); err != nil {
		return ServiceConfig{}, err
	}
	return cfg, nil
}
