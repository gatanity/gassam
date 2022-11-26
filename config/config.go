package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/aws/aws-sdk-go/aws/defaults"
	"gopkg.in/ini.v1"
)

// Config is this tool's configuration
type Config struct {
	Subdomain                   string
	OneloginAppId               string
	AwsAccountId                string
	DefaultSessionDurationHours int
	ChromeUserDataDir           string
}

const (
	subdomainKeyName                   = "subdomain"
	oneloginAppIdKeyName               = "onelogin_app_id"
	awsAccountIdKeyName                = "aws_account_id"
	defaultSessionDurationHoursKeyName = "default_session_duration_hours"
	chromeUserDataDirKeyName           = "chrome_user_data_dir"
)

// NewConfig returns Config from default AWS config file
func NewConfig(profile string) (Config, error) {
	cfg := Config{}

	f, err := loadConfigFile()
	if err != nil {
		return cfg, err
	}

	section, err := f.GetSection(sectionName(profile))
	if err != nil {
		return cfg, err
	}

	subdomainKey, err := section.GetKey(subdomainKeyName)
	if err != nil {
		return cfg, err
	}

	oneloginAppIdKey, err := section.GetKey(oneloginAppIdKeyName)
	if err != nil {
		return cfg, err
	}

	awsAccountIdKey, err := section.GetKey(awsAccountIdKeyName)
	if err != nil {
		return cfg, err
	}

	defaultSessionDurationHoursKey, err := section.GetKey(defaultSessionDurationHoursKeyName)
	if err != nil {
		return cfg, err
	}

	userDataDirKey, err := section.GetKey(chromeUserDataDirKeyName)
	if err != nil {
		return cfg, err
	}

	cfg.Subdomain = subdomainKey.Value()
	cfg.OneloginAppId = oneloginAppIdKey.Value()
	cfg.AwsAccountId = awsAccountIdKey.Value()
	defaultSessionDurationHours, err := strconv.Atoi(defaultSessionDurationHoursKey.Value())
	if err != nil {
		return cfg, err
	}
	cfg.DefaultSessionDurationHours = defaultSessionDurationHours
	cfg.ChromeUserDataDir = userDataDirKey.Value()

	return cfg, nil
}

// Save saves config to file.
func Save(cfg Config, profile string) error {
	f, err := loadConfigFile()
	if err != nil {
		return err
	}

	section := f.Section(sectionName(profile))

	section.Key(subdomainKeyName).SetValue(cfg.Subdomain)
	section.Key(oneloginAppIdKeyName).SetValue(cfg.OneloginAppId)
	section.Key(awsAccountIdKeyName).SetValue(cfg.AwsAccountId)
	section.Key(defaultSessionDurationHoursKeyName).SetValue(strconv.Itoa(cfg.DefaultSessionDurationHours))
	section.Key(chromeUserDataDirKeyName).SetValue(cfg.ChromeUserDataDir)

	file := getConfigFilename()
	dir := filepath.Dir(file)
	err = os.MkdirAll(dir, os.FileMode(0755))
	if err != nil {
		return err
	}

	return f.SaveTo(file)
}

func getConfigFilename() string {
	// https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-envvars.html
	file := os.Getenv("AWS_CONFIG_FILE")
	if len(file) == 0 {
		file = defaults.SharedConfigFilename()
	}
	return file
}

func loadConfigFile() (*ini.File, error) {
	file := getConfigFilename()
	return ini.LooseLoad(file)
}

func sectionName(profile string) string {
	if profile == "default" {
		return profile
	}
	return fmt.Sprintf("profile %s", profile)
}
