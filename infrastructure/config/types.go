package config

var Service = "poly-kit"
var Version = "v1.0.0"
var GitCommit string
var OSBuildName string
var BuildDate string

type MainConfig struct {
	Log struct {
		Level  string `yaml:"level" validate:"oneof=trace debug info warn error fatal panic"`
		Format string `yaml:"format" validate:"oneof=text json"`
	} `yaml:"log"`
	Server struct {
		Listen string `yaml:"listen"  validate:"hostname_port"`
	} `yaml:"server"`
	APM struct {
		Enabled bool     `yaml:"enabled"`
		Host    string   `yaml:"host" validate:"hostname"`
		Port    int      `yaml:"port" validate:"required,min=1,max=65535"`
		Rate    *float64 `yaml:"rate" validate:"omitempty,min=0.1,max=1"`
	} `yaml:"apm"`
	App struct {
		Name    string `yaml:"name" validate:"required"`
		Version string `yaml:"version" validate:"required"`
		Env     string `yaml:"env" validate:"required"`
		Tribe   string `yaml:"tribe" validate:"required"`
	}
}
