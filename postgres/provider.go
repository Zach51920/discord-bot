package postgres

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"log/slog"
	"time"
)

type Config struct {
	PostgresURL string
}

type Provider struct {
	config     Config
	db         *sqlx.DB
	shutdownCh chan struct{}
}

func NewProvider(c Config) (*Provider, error) {
	db, err := getPostgresConnection(c)
	if err != nil {
		return nil, fmt.Errorf("get connection: %w", err)
	}
	p := &Provider{db: db, shutdownCh: make(chan struct{})}
	go p.keepAlive()
	return p, nil
}

func (p *Provider) Get() *sqlx.DB {
	return p.db
}

func getPostgresConnection(c Config) (*sqlx.DB, error) {
	if c.PostgresURL == "" {
		return nil, errors.New("postgres url is null")
	}

	db, err := sql.Open("postgres", c.PostgresURL)
	if err != nil {
		return nil, fmt.Errorf("open: %w", err)
	}

	// verify connection
	for i := 0; i < 10; i++ {
		if err = db.Ping(); err != nil {
			break
		}
		time.Sleep(1 * time.Second)
	}
	if err = db.Ping(); err != nil {
		return nil, errors.New("max connection attempts exceeded")
	}

	// set connection options
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(3)
	db.SetConnMaxLifetime(3 * time.Second)

	return sqlx.NewDb(db, "postgres"), nil
}

func (p *Provider) keepAlive() {
	ticker := time.NewTicker(time.Minute)
	for {
		select {
		case <-p.shutdownCh:
			return
		case <-ticker.C:
			if err := p.db.Ping(); err != nil {
				slog.Warn("failed to ping postgres", "error", err)
			} else {
				slog.Debug("postgres ping successful")
			}
		}
	}
}

func (p *Provider) Close() error {
	close(p.shutdownCh)
	return p.db.Close()
}
