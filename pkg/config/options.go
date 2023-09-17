package config

type Options struct {
	User           string
	Password       string
	Endpoint       string
	Table          string
	KubeConfigPath string
	Port           int
	HealthPort     int
	DebugMode      bool
}
