package http_config

type HTTPConfig struct {
	Addr string `yaml:"addr" env_default:"0.0.0.0"`
	Port string `yaml:"port" env_default:"8080"`
}
