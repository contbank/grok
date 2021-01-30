package grok

import (
	"io/ioutil"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

//Settings ...
type Settings struct {
	API          *APISettings   `yaml:"api"`
	Mongo        *MongoSettings `yaml:"mongo"`
	UserProvider *UserProvider  `yaml:"user_provider"`
	Mail         *MailSettings  `yaml:"mail"`
	AWS          *AWSSettings   `yaml:"aws"`
}

// APISettings ...
type APISettings struct {
	Host    string   `yaml:"host"`
	Swagger string   `yaml:"swagger"`
	Auth    *APIAuth `yaml:"auth"`
}

// MongoSettings ...
type MongoSettings struct {
	CaFilePath       *string `yaml:"ca_file_path"`
	ConnectionString string  `yaml:"connection_string"`
	Database         string  `yaml:"database"`
}

// AWSSettings ...
type AWSSettings struct {
	SNS *AWSCredentials `yaml:"sns"`
	SQS *AWSCredentials `yaml:"sqs"`
	IAM *AWSCredentials `yaml:"iam"`
	S3  *AWSCredentials `yaml:"s3"`
}

//AWSCredentials ...
type AWSCredentials struct {
	Path     string `yaml:"path"`
	Fake     bool   `yaml:"fake"`
	Endpoint string `yaml:"endpoint"`
	Region   string `yaml:"region"`
}

// APIAuth ...
type APIAuth struct {
	Fake       bool         `yaml:"fake"`
	FakeConfig *FakeAPIAuth `yaml:"fake_config"`
	Tenant     string       `yaml:"tenant"`
	JWKS       string       `yaml:"jwks"`
	Audience   []string     `yaml:"audience"`
}

// FakeAPIAuth ...
type FakeAPIAuth struct {
	Claims        map[string]interface{} `yaml:"claims"`
	Authenticated bool                   `yaml:"authenticated"`
}

// UserProvider ...
type UserProvider struct {
	Kind  string                   `yaml:"kind"`
	Mock  []map[string]interface{} `yaml:"users"`
	Auth0 struct {
		CacheTTL        int64  `yaml:"cache_ttl"`
		Domain          string `yaml:"domain"`
		ClientFrom      string `yaml:"client_from"`
		ClientID        string `yaml:"client_id"`
		ClientSecret    string `yaml:"client_secret"`
		ClientIDEnv     string `yaml:"client_id_env"`
		ClientSecretEnv string `yaml:"client_secret_env"`
	}
}

// MailSettings ...
type MailSettings struct {
	Provider string `yaml:"provider"`
	Fake     struct {
		ShouldReturnError bool `yaml:"should_return_error"`
	} `yaml:"fake"`
	SendGrid struct {
		APIKey    string `yaml:"api_key"`
		FromEnv   bool   `yaml:"from_env"`
		APIKeyEnv string `yaml:"api_key_env"`
	} `yaml:"send_grid"`
}

// FromYAML ...
func FromYAML(file string, dist interface{}) error {
	filename, _ := filepath.Abs(file)

	data, err := ioutil.ReadFile(filename)

	if err != nil {
		return err
	}

	return yaml.Unmarshal(data, dist)
}
