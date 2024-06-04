package environment

type DbEnv struct {
	Host     string
	User     string
	Password string
	DbName   string
	Port     int64
}

func ReadDb() (*DbEnv, error) {
	host, err := readString("POSTGRES_HOST", "localhost", false)
	if err != nil {
		return nil, err
	}

	user, err := readString("POSTGRES_USER", "postgres", false)
	if err != nil {
		return nil, err
	}

	password, err := readString("POSTGRES_PASSWORD", "mysecretpassword", false)
	if err != nil {
		return nil, err
	}

	dbName, err := readString("POSTGRES_DB_NAME", "postgres", false)
	if err != nil {
		return nil, err
	}

	port, err := readInt("POSTGRES_PORT", 5432, false)
	if err != nil {
		return nil, err
	}

	return &DbEnv{
		Host:     *host,
		User:     *user,
		Password: *password,
		DbName:   *dbName,
		Port:     *port,
	}, nil
}
