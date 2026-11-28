# checkers

A minimal Go implementation of checkers to learn goroutines and worker pools in a concurrent environment, without letting it hog CPU/memory.

go get github.com/jackc/pgx/v5/pgxpool
go get github.com/jackc/pgx/v5/stdlib

CREATE TABLE compact_states (
    id     BIGINT   NOT NULL,
    whodo  SMALLINT NOT NULL,
    cost   INTEGER  NOT NULL,
    PRIMARY KEY (id, whodo)
);

CREATE TABLE states_relations (
    id            BIGINT   NOT NULL,  -- id дочернего состояния
    id_whodo      SMALLINT NOT NULL,  -- whodo дочернего состояния
    parent_id     BIGINT   NOT NULL,  -- id родительского состояния
    parent_whodo  SMALLINT NOT NULL,  -- whodo родительского состояния
    PRIMARY KEY (id, id_whodo, parent_id, parent_whodo)
);



