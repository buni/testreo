//nolint:gochecknoglobals
package dt

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/buni/wallet/internal/pkg/testing/migrate"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats-server/v2/server"
	tnats "github.com/nats-io/nats-server/v2/test"
	"github.com/nats-io/nats.go"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)

var (
	DB       *pgxpool.Pool
	NATSConn *nats.Conn
)

func SetupNATS() {
	var err error
	go tnats.RunServer(&server.Options{ //nolint
		Port:      4229,
		JetStream: true,
	}).Start()

	NATSConn, err = nats.Connect("nats://localhost:4229")
	if err != nil {
		log.Fatalln("failed to connect to nats", err)
	}
}

func SetupPostgres() *dockertest.Resource {
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalln("failed to create dt pool", err)
	}

	password := "12345678"
	user := "pgtestuser"
	database := "pgtestdb"

	runOptions := &dockertest.RunOptions{ //nolint
		Repository: "postgres",
		Tag:        "latest",
		Env: []string{
			"POSTGRES_PASSWORD=" + password,
			"POSTGRES_USER=" + user,
			"POSTGRES_DB=" + database,
			"listen_addresses = '*'",
		},
	}

	resource, err := pool.RunWithOptions(runOptions, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"} //nolint
	})
	if err != nil {
		log.Fatalln("could not start database resource", err)
	}

	hostAndPort := resource.GetHostPort("5432/tcp")

	databaseURL := fmt.Sprintf(
		"postgres://%s:%s@%s/%s?sslmode=disable",
		user,
		password,
		hostAndPort,
		database,
	)

	err = resource.Expire(600)
	if err != nil {
		log.Fatalln("could not set expire time for database resource", err)
	}

	pool.MaxWait = 60 * time.Second

	err = pool.Retry(func() (err error) {
		conf, err := pgxpool.ParseConfig(databaseURL)
		if err != nil {
			log.Println("failed to parse config", err)
			return fmt.Errorf("failed to parse config: %w", err)
		}

		DB, err = pgxpool.NewWithConfig(context.Background(), conf)
		if err != nil {
			log.Println("failed to create pgxpool", err)
			return fmt.Errorf("failed to create pgxpool: %w", err)
		}

		err = DB.Ping(context.Background())
		if err != nil {
			log.Println("failed to ping database, retrying ...", err)
			return fmt.Errorf("failed to ping database: %w", err)
		}

		return nil
	})
	if err != nil {
		log.Fatalln("failed to connect to docker", err)
	}

	err = migrate.Migrate(databaseURL, "../../../../migrations")
	if err != nil {
		log.Fatalln("failed to migrate db", err)
	}

	return resource
}

// Cleanup destroys the ephemeral test database.
func Cleanup(resource *dockertest.Resource) error {
	if err := resource.Close(); err != nil {
		return fmt.Errorf("could not close resource: %w", err)
	}
	return nil
}
