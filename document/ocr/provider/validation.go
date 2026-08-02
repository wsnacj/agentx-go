package provider

import (
	"fmt"
	"strings"

	"github.com/wsnacj/agentx-go/document/ocr/config"
)

// ConfigValidator 用于在构建 Provider 前补全和校验配置。
type ConfigValidator func(*config.ProviderConfig) error

// DefaultConfigValidators 返回内建 Provider 的配置校验器。
func DefaultConfigValidators() map[string]ConfigValidator {
	return map[string]ConfigValidator{
		"textin": func(cfg *config.ProviderConfig) error {
			if cfg.Auth == nil {
				cfg.Auth = map[string]string{}
			}
			appID := strings.TrimSpace(cfg.Auth["app_id"])
			secret := strings.TrimSpace(cfg.Auth["secret_code"])
			if appID == "" {
				return fmt.Errorf("textin provider requires auth.app_id")
			}
			if secret == "" {
				return fmt.Errorf("textin provider requires auth.secret_code")
			}
			cfg.Auth["app_id"] = appID
			cfg.Auth["secret_code"] = secret
			if err := validateTextInHeaders(cfg.Headers); err != nil {
				return err
			}
			return nil
		},
		"volcengine": func(cfg *config.ProviderConfig) error {
			if cfg.Auth == nil {
				cfg.Auth = map[string]string{}
			}
			accessKey := strings.TrimSpace(cfg.Auth["access_key_id"])
			secretKey := strings.TrimSpace(cfg.Auth["secret_access_key"])
			if accessKey == "" {
				return fmt.Errorf("volcengine provider requires auth.access_key_id")
			}
			if secretKey == "" {
				return fmt.Errorf("volcengine provider requires auth.secret_access_key")
			}
			cfg.Auth["access_key_id"] = accessKey
			cfg.Auth["secret_access_key"] = secretKey
			if token := strings.TrimSpace(cfg.Auth["security_token"]); token != "" {
				cfg.Auth["security_token"] = token
			}
			if cfg.Additional == nil {
				cfg.Additional = map[string]any{}
			}
			if _, ok := cfg.Additional["action"]; !ok {
				cfg.Additional["action"] = defaultVolcAction
			}
			if _, ok := cfg.Additional["version"]; !ok {
				cfg.Additional["version"] = defaultVolcVersion
			}
			if _, ok := cfg.Additional["region"]; !ok {
				cfg.Additional["region"] = defaultVolcRegion
			}
			if _, ok := cfg.Additional["service"]; !ok {
				cfg.Additional["service"] = defaultVolcService
			}
			return nil
		},
		"baidu": func(cfg *config.ProviderConfig) error {
			if cfg.BaseURL == "" {
				cfg.BaseURL = "https://aip.baidubce.com/rest/2.0/ocr/v1/accurate"
			}
			if cfg.Auth == nil {
				cfg.Auth = map[string]string{}
			}
			accessToken := strings.TrimSpace(cfg.Auth["access_token"])
			apiKey := strings.TrimSpace(cfg.Auth["api_key"])
			secretKey := strings.TrimSpace(cfg.Auth["secret_key"])
			if accessToken == "" {
				if apiKey == "" || secretKey == "" {
					return fmt.Errorf("baidu provider requires either auth.access_token or auth.api_key/auth.secret_key")
				}
			}
			cfg.Auth["access_token"] = accessToken
			cfg.Auth["api_key"] = apiKey
			cfg.Auth["secret_key"] = secretKey
			if cfg.Additional == nil {
				cfg.Additional = map[string]any{}
			}
			if _, ok := cfg.Additional["token_url"]; !ok {
				cfg.Additional["token_url"] = "https://aip.baidubce.com/oauth/2.0/token"
			}
			return nil
		},
	}
}
