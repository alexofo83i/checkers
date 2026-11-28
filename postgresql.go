package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
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

// getStatesByParent загружает дочерние состояния state из БД.
// whodo — «чей ход» у дочерних состояний (для checkers всегда
// противоположен whodo родителя), используется как дополнительный
// фильтр и страховка от несогласованных данных.
//
// Требуемая схема:
//
//	CREATE TABLE compact_states (
//	    id     BIGINT   NOT NULL,
//	    whodo  SMALLINT NOT NULL,
//	    cost   INTEGER  NOT NULL,
//	    PRIMARY KEY (id, whodo)
//	);
//
//	CREATE TABLE states_relations (
//	    id            BIGINT   NOT NULL,  -- id дочернего состояния
//	    id_whodo      SMALLINT NOT NULL,  -- whodo дочернего состояния
//	    parent_id     BIGINT   NOT NULL,  -- id родительского состояния
//	    parent_whodo  SMALLINT NOT NULL,  -- whodo родительского состояния
//	    PRIMARY KEY (id, id_whodo, parent_id, parent_whodo)
//	);
func getStatesByParent(state *PlayState, whodo byte) []*PlayState {
	if state == nil || pool == nil {
		return nil
	}

	ctx := context.Background()
	parentCS := state.packState() // parentCS.ID, parentCS.Whodo

	rows, err := pool.Query(ctx, `
		SELECT cs.id, cs.whodo, cs.cost
		FROM   states_relations r
		JOIN   compact_states   cs
		       ON cs.id = r.id AND cs.whodo = r.id_whodo
		WHERE  r.parent_id    = $1
		  AND  r.parent_whodo = $2
		  AND  r.id_whodo     = $3
	`, parentCS.ID, int16(parentCS.Whodo), int16(whodo))
	if err != nil {
		log.Printf("getStatesByParent: query failed: %v", err)
		return nil
	}
	defer rows.Close()

	var children []*PlayState
	for rows.Next() {
		var (
			id   int64
			w    int16
			cost int
		)
		if err := rows.Scan(&id, &w, &cost); err != nil {
			log.Printf("getStatesByParent: scan failed: %v", err)
			continue
		}

		place, color := DecodeID(id)
		child := (&CompactState{
			ID:    id,
			Whodo: uint8(w),
			Cost:  cost,
			Place: place,
			Color: color,
		}).unpackState()

		child.ID = id
		child.level = state.level + 1
		child.prevState = state

		children = append(children, child)
		state.nextStates = append(state.nextStates, child)
	}
	if err := rows.Err(); err != nil {
		log.Printf("getStatesByParent: rows error: %v", err)
	}
	return children
}

// putStatesByParent сохраняет состояния states как дочерние для state
// и заводит связи в states_relations. Перед вставкой удаляет прежние
// дочерние записи (только «сирот» — тех, кто больше ни с кем не связан)
// и все старые связи этого родителя. Всё — в одной транзакции, без MERGE.
// whodo — whodo дочерних состояний.
func putStatesByParent(state *PlayState, whodo byte, states []*PlayState) {
	if state == nil || pool == nil {
		return
	}

	ctx := context.Background()
	parentCS := state.packState() // parentCS.ID, parentCS.Whodo

	tx, err := pool.Begin(ctx)
	if err != nil {
		log.Printf("putStatesByParent: begin failed: %v", err)
		return
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback(ctx)
		}
	}()

	// 1) Удаляем «сирот» — детей этого родителя, которые больше ни
	//    с кем не связаны. «Наш» родитель однозначно определяется
	//    парой (parent_id, parent_whodo).
	if _, err := tx.Exec(ctx, `
		DELETE FROM compact_states cs
		WHERE (cs.id, cs.whodo) IN (
		    SELECT id, id_whodo
		    FROM   states_relations
		    WHERE  parent_id    = $1
		      AND  parent_whodo = $2
		      AND  id_whodo     = $3
		)
		AND NOT EXISTS (
		    SELECT 1 FROM states_relations r2
		    WHERE  r2.id         = cs.id
		      AND  r2.id_whodo   = cs.whodo
		      AND  NOT (r2.parent_id    = $1
		            AND r2.parent_whodo = $2)
		)
	`, parentCS.ID, int16(parentCS.Whodo), int16(whodo)); err != nil {
		log.Printf("putStatesByParent: delete orphan children failed: %v", err)
		return
	}

	// 2) Удаляем все старые связи этого родителя.
	if _, err := tx.Exec(ctx, `
		DELETE FROM states_relations
		WHERE parent_id = $1 AND parent_whodo = $2
	`, parentCS.ID, int16(parentCS.Whodo)); err != nil {
		log.Printf("putStatesByParent: delete relations failed: %v", err)
		return
	}

	// 3) Вставляем родителя, детей и связи.
	batch := &pgx.Batch{}

	// Родительская строка: при повторной вставке обновляем cost,
	// чтобы поддерживать консистентность без MERGE.
	batch.Queue(`
		INSERT INTO compact_states (id, whodo, cost)
		VALUES ($1, $2, $3)
		ON CONFLICT (id, whodo) DO UPDATE SET cost = EXCLUDED.cost
	`, parentCS.ID, int16(parentCS.Whodo), parentCS.Cost)

	for _, child := range states {
		if child == nil {
			continue
		}
		childCS := child.packState()
		childCS.Whodo = whodo // фиксируем по параметру сигнатуры

		batch.Queue(`
			INSERT INTO compact_states (id, whodo, cost)
			VALUES ($1, $2, $3)
			ON CONFLICT (id, whodo) DO UPDATE SET cost = EXCLUDED.cost
		`, childCS.ID, int16(childCS.Whodo), childCS.Cost)

		batch.Queue(`
			INSERT INTO states_relations
			       (id, id_whodo, parent_id, parent_whodo)
			VALUES ($1, $2,       $3,        $4)
			ON CONFLICT DO NOTHING
		`, childCS.ID, int16(childCS.Whodo), parentCS.ID, int16(parentCS.Whodo))
	}

	br := tx.SendBatch(ctx, batch)
	for i := 0; i < batch.Len(); i++ {
		if _, err := br.Exec(); err != nil {
			_ = br.Close()
			log.Printf("putStatesByParent: batch exec [%d] failed: %v", i, err)
			return
		}
	}
	if err := br.Close(); err != nil {
		log.Printf("putStatesByParent: batch close failed: %v", err)
		return
	}

	if err := tx.Commit(ctx); err != nil {
		log.Printf("putStatesByParent: commit failed: %v", err)
		return
	}
	committed = true
}
