package store

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	ConfigBase           = "base"
	ConfigKmsKey         = "kms-key"
	ConfigAwsRegion      = "aws-region"
	ConfigUsePaths       = "use-paths"
	ConfigSkipBaseConfig = "skip-base-config"
	ConfigRetries        = "retries"
)

// KMS key ID seems to be an RFC 4122
var looksLikeKmsKeyId = regexp.MustCompile("^[A-Za-z0-9]{8}-[A-Za-z0-9]{4}-[A-Za-z0-9]{4}-[A-Za-z0-9]{4}-[A-Za-z0-9]{12}$")

type Config struct {
	vp *viper.Viper
}

func NewConfig() *Config {
	c := Config{viper.New()}
	c.vp.SetTypeByDefaultValue(true)

	c.vp.BindEnv(ConfigBase, "CHAMBER_BASE")

	c.vp.SetDefault(ConfigKmsKey, "alias/parameter_store_key")
	if _, ok := os.LookupEnv("CHAMBER_KMS_KEY"); ok {
		c.vp.BindEnv(ConfigKmsKey, "CHAMBER_KMS_KEY")
	} else {
		c.vp.BindEnv(ConfigKmsKey, "CHAMBER_KMS_KEY_ALIAS")
	}

	c.vp.BindEnv(ConfigAwsRegion, "CHAMBER_AWS_REGION")

	c.vp.SetDefault(ConfigUsePaths, false)
	c.vp.BindEnv(ConfigUsePaths, "CHAMBER_USE_PATHS")

	c.vp.SetDefault(ConfigSkipBaseConfig, false)
	c.vp.BindEnv(ConfigSkipBaseConfig, "CHAMBER_SKIP_BASE_CONFIG")

	c.vp.SetDefault(ConfigRetries, 10)
	c.vp.BindEnv(ConfigRetries, "CHAMBER_RETRIES")

	return &c
}

func (c *Config) BindPFlag(name string, flag *pflag.Flag) error {
	return c.vp.BindPFlag(name, flag)
}

func (c *Config) Base() (v string, ok bool) {
	i, ok := c.getConfigValOptional(ConfigBase)
	if ok {
		v = i.(string)
		ok = v != ""
	}
	return
}

func (c *Config) KmsKey() string {
	key := c.getConfigValRequired(ConfigKmsKey).(string)
	if strings.HasPrefix(key, "arn:aws:kms:") {
		return key
	} else if looksLikeKmsKeyId.MatchString(key) {
		return key
	} else if !strings.HasPrefix(key, "alias/") {
		return fmt.Sprintf("alias/%s", key)
	}
	return key
}

func (c *Config) AwsRegion() (v string, ok bool) {
	i, ok := c.getConfigValOptional(ConfigAwsRegion)
	if ok {
		v = i.(string)
	}
	return
}

func (c *Config) UsePaths() bool {
	return c.getConfigValRequired(ConfigUsePaths).(bool)
}

func (c *Config) BaseConfigPath() string {
	base, ok := c.Base()
	if c.getConfigValRequired(ConfigSkipBaseConfig).(bool) || !ok {
		return ""
	}
	return base
}

func (c *Config) Retries() int {
	return c.getConfigValRequired(ConfigRetries).(int)
}

func (c *Config) MergeConfig(config string) error {
	if strings.HasPrefix(config, "{") {
		c.vp.SetConfigType("json")
	} else {
		c.vp.SetConfigType("properties")
	}
	return c.vp.MergeConfig(strings.NewReader(config))
}

func (c *Config) String() string {
	return fmt.Sprintf("%v", c.vp.AllSettings())
}

func (c *Config) Marshal() (string, error) {
	b, err := json.MarshalIndent(c.vp.AllSettings(), "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (c *Config) getConfigVal(name string, required bool) (interface{}, bool) {
	if c.vp.IsSet(name) {
		return c.vp.Get(name), true
	} else {
		if required {
			panic(fmt.Sprintf("%s config option should have had a default", name))
		} else {
			return nil, false
		}
	}
}

func (c *Config) getConfigValOptional(name string) (interface{}, bool) {
	return c.getConfigVal(name, false)
}

func (c *Config) getConfigValRequired(name string) interface{} {
	v, _ := c.getConfigVal(name, true)
	return v
}
