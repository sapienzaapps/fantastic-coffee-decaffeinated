package main

import (
	"errors"
	"fmt"
	"github.com/ardanlabs/conf/v3"
	"gopkg.in/yaml.v3"
	"io"
	"os"
	"time"
)

// WebAPIConfiguration describes the web API configuration. This structure is automatically parsed by
// loadConfiguration and values from flags, environment variable or configuration file will be loaded.
type WebAPIConfiguration struct {
	Config struct {
		Path string `conf:"default:/conf/config.yml"`
	}
	Web struct {
		APIHost         string        `conf:"default:0.0.0.0:3000"`
		DebugHost       string        `conf:"default:0.0.0.0:4000"`
		ReadTimeout     time.Duration `conf:"default:5s"`
		WriteTimeout    time.Duration `conf:"default:5s"`
		ShutdownTimeout time.Duration `conf:"default:5s"`
	}
	Debug bool
	DB    struct {
		// Driver selects the database engine: "sqlite" or "postgres".
		Driver string `conf:"default:sqlite"`
		// Filename is the SQLite database file (used only when Driver is "sqlite").
		Filename string `conf:"default:/tmp/decaf.db"`
		// DSN is the PostgreSQL connection string (used only when Driver is "postgres"), e.g.
		// postgres://user:password@localhost:5432/decaf?sslmode=disable
		DSN string `conf:"mask"`
	}
	// S3 configures the (optional) S3-compatible object storage. It is enabled only when Endpoint is set; then all
	// the other fields are required.
	// The env and flag names are set explicitly because the conf library would otherwise split "S3" into "S_3".
	S3 struct {
		// Endpoint is the S3 API URL, including the scheme (http:// or https://), e.g. http://localhost:3900. It is
		// used both by the backend and for generating presigned URLs.
		Endpoint string `conf:"env:S3_ENDPOINT,flag:s3-endpoint"`
		// Region is used for signing requests (for Garage, it is the s3_region value in garage.toml).
		Region string `conf:"env:S3_REGION,flag:s3-region"`
		// Bucket is the name of the bucket where files are stored.
		Bucket string `conf:"env:S3_BUCKET,flag:s3-bucket"`
		// AccessKeyID and SecretAccessKey are the credentials of the access key.
		AccessKeyID     string `conf:"env:S3_ACCESS_KEY_ID,flag:s3-access-key-id"`
		SecretAccessKey string `conf:"mask,env:S3_SECRET_ACCESS_KEY,flag:s3-secret-access-key"`
	}
}

// loadConfiguration creates a WebAPIConfiguration starting from flags, environment variables and configuration file.
// It works by loading environment variables first, then update the config using command line flags, finally loading the
// configuration file (specified in WebAPIConfiguration.Config.Path).
// So, CLI parameters will override the environment, and configuration file will override everything.
// Note that the configuration file can be specified only via CLI or environment variable.
func loadConfiguration() (WebAPIConfiguration, error) {
	var cfg WebAPIConfiguration

	// Try to load configuration from environment variables and command line switches
	usage, err := conf.Parse("CFG", &cfg)
	if err != nil {
		if errors.Is(err, conf.ErrHelpWanted) {
			fmt.Println(usage) //nolint:forbidigo
			return cfg, conf.ErrHelpWanted
		}
		return cfg, fmt.Errorf("parsing config: %w", err)
	}

	// Override values from YAML if specified and if it exists (useful in k8s/compose)
	fp, err := os.Open(cfg.Config.Path)
	if err != nil && !os.IsNotExist(err) {
		return cfg, fmt.Errorf("can't read the config file, while it exists: %w", err)
	} else if err == nil {
		yamlFile, err := io.ReadAll(fp)
		if err != nil {
			return cfg, fmt.Errorf("can't read config file: %w", err)
		}
		err = yaml.Unmarshal(yamlFile, &cfg)
		if err != nil {
			return cfg, fmt.Errorf("can't unmarshal config file: %w", err)
		}
		_ = fp.Close()
	}

	return cfg, nil
}
