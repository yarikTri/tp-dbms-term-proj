package configs

type postgresConfig struct {
	User     string
	Password string
	DB       string
	Port     string
}

var PostgresConfig postgresConfig

func init() {
	PostgresConfig = postgresConfig{
		User:     "docker",
		Password: "docker",
		DB:       "docker",
		Port:     "5432",
	}
}
