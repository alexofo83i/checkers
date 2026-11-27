package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib" // для database/sql, если где-то нужен *sql.DB
)

// ─────────────────────────────────────────────────────────────────────────────
// Параметры подключения. Переопределяются при необходимости до InitDB().
// ─────────────────────────────────────────────────────────────────────────────
var (
	// Подключение
	dbHost     = "localhost"
	dbPort     = 5432
	dbUser     = "postgres"
	dbPassword = "postgres"
	dbName     = "checkers"
	dbSSLMode  = "disable"       // disable | require | verify-full | ...
	dbTimeout  = 5 * time.Second // таймаут на подключение/запросы в пуле

	// Настройки пула
	dbPoolMaxConns          int32 = 10
	dbPoolMinConns          int32 = 0
	dbPoolMaxConnLifetime         = time.Hour
	dbPoolMaxConnIdleTime         = 30 * time.Minute
	dbPoolHealthCheckPeriod       = time.Minute
)

// Глобальный пул, используется getStatesByParent / putStatesByParent.
var pool *pgxpool.Pool

// Возвращает DSN (connection string) из текущих параметров.
func buildDSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s connect_timeout=%d",
		dbHost, dbPort, dbUser, dbPassword, dbName, dbSSLMode,
		int(dbTimeout.Seconds()),
	)
}

// InitDB инициализирует пул подключений к PostgreSQL.
// Должна быть вызвана один раз при старте приложения.
// Повторный вызов при уже инициализированном пуле — no-op.
func InitDB() error {
	if pool != nil {
		return nil
	}

	cfg, err := pgxpool.ParseConfig(buildDSN())
	if err != nil {
		return fmt.Errorf("postgres: parse config: %w", err)
	}

	if dbPoolMaxConns > 0 {
		cfg.MaxConns = dbPoolMaxConns
	}
	cfg.MinConns = dbPoolMinConns
	cfg.MaxConnLifetime = dbPoolMaxConnLifetime
	cfg.MaxConnIdleTime = dbPoolMaxConnIdleTime
	cfg.HealthCheckPeriod = dbPoolHealthCheckPeriod
	cfg.ConnConfig.ConnectTimeout = dbTimeout

	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	p, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return fmt.Errorf("postgres: create pool: %w", err)
	}

	// Проверяем, что база действительно доступна.
	if err := p.Ping(ctx); err != nil {
		p.Close()
		return fmt.Errorf("postgres: ping: %w", err)
	}

	pool = p
	log.Printf("postgres: pool initialized (host=%s port=%d db=%s maxConns=%d)",
		dbHost, dbPort, dbName, cfg.MaxConns)
	return nil
}

// CloseDB закрывает пул. Безопасно вызывать многократно и при pool == nil.
func CloseDB() {
	if pool == nil {
		return
	}
	pool.Close()
	pool = nil
	log.Println("postgres: pool closed")
}

// DB возвращает *sql.DB поверх того же пула — на случай, если понадобится
// стандартный database/sql интерфейс (например, для sqlx или сторонних утилит).
// Не закрывайте возвращённый *sql.DB самостоятельно — управление у пула.
func DB() *sql.DB {
	if pool == nil {
		return nil
	}
	db := stdlib.OpenDBFromPool(pool)
	return db
}

// PingDB — удобная проверка живости БД (например, в health-check).
func PingDB(ctx context.Context) error {
	if pool == nil {
		return fmt.Errorf("postgres: pool is not initialized")
	}
	return pool.Ping(ctx)
}
