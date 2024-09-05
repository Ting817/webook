package cfg

type Config struct {
	K8s struct {
		Addr      string `toml:"addr"`
		Token     string `toml:"token"`
		Namespace string `toml:"namespace"`
	} `toml:"k8s"`
	DB struct {
		DSN string `toml:"dsn"`
	} `toml:"db"`
	Redis struct {
		Addr string `toml:"addr"`
	} `toml:"redis"`
}
