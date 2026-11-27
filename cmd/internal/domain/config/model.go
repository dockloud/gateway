package config

type MergedConfig struct {
	Name         string
	ID           string
	Port         string
	Timeout      int
	PublicRoutes []string
	JWT          JWTConfig
	Dependencies map[string]DependencyConfig
}

type DependencyConfig struct {
	Name    string
	Address string
}

type JWTConfig struct {
	ServiceSecret []byte
	UserAuthToken string
}
