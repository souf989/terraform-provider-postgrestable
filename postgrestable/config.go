package postgresql

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/lib/pq" //PostgreSQL db
	"gocloud.dev/postgres"
	_ "gocloud.dev/postgres/awspostgres"
	"net/url"
	"strconv"
	"strings"
	"sync"
)

var (
	dbRegistryLock sync.Mutex
	dbRegistry     map[string]*DBConnection = make(map[string]*DBConnection, 1)
)

type DBConnection struct {
	*sql.DB
	client *Client
}

// Config - provider config
type Config struct {
	Scheme            string
	Host              string
	Port              int
	Username          string
	Password          string
	SSLMode           string
	Timeout           int
	ConnectTimeoutSec int
	MaxConns          int
}

// Client struct holding connection string
type Client struct {
	// Configuration for the client
	config Config

	databaseName string
}

// NewClient returns client config for the specified database.
func (c *Config) NewClient(database string) *Client {
	return &Client{
		config:       *c,
		databaseName: database,
	}
}

func (c *Config) connParams() []string {
	params := map[string]string{}

	// sslmode and connect_timeout are not allowed with gocloud
	// (TLS is provided by gocloud directly)
	if c.Scheme == "postgres" {
		params["sslmode"] = c.SSLMode
		params["connect_timeout"] = strconv.Itoa(c.ConnectTimeoutSec)
	}

	paramsArray := []string{}
	for key, value := range params {
		paramsArray = append(paramsArray, fmt.Sprintf("%s=%s", key, url.QueryEscape(value)))
	}

	return paramsArray
}

func (c *Config) connStr(database string) string {
	host := c.Host

	connStr := fmt.Sprintf(
		"%s://%s:%s@%s:%d/%s?%s",
		c.Scheme,
		url.QueryEscape(c.Username),
		url.QueryEscape(c.Password),
		host,
		c.Port,
		database,
		strings.Join(c.connParams(), "&"),
	)

	return connStr
}

// Connect returns a copy to an sql.Open()'ed database connection wrapped in a DBConnection struct.
// Callers must return their database resources. Use of QueryRow() or Exec() is encouraged.
// Query() must have their rows.Close()'ed.
func (c *Client) Connect() (*DBConnection, error) {
	dbRegistryLock.Lock()
	defer dbRegistryLock.Unlock()

	dsn := c.config.connStr(c.databaseName)
	conn, found := dbRegistry[dsn]
	if !found {

		var db *sql.DB
		var err error
		if c.config.Scheme == "postgres" {
			db, err = sql.Open("postgres", dsn)
		} else {
			db, err = postgres.Open(context.Background(), dsn)
		}
		if err != nil {
			return nil, fmt.Errorf("Error connecting to PostgreSQL server %s (scheme: %s): %w", c.config.Host, c.config.Scheme, err)
		}

		// We don't want to retain connection
		// So when we connect on a specific database which might be managed by terraform,
		// we don't keep opened connection in case of the db has to be dopped in the plan.
		db.SetMaxIdleConns(0)
		db.SetMaxOpenConns(c.config.MaxConns)

		conn = &DBConnection{
			db,
			c,
		}
		dbRegistry[dsn] = conn
	}

	return conn, nil
}
