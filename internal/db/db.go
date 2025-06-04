package db

import (
	"context"
	"fmt"
	"time"

	"github.com/cockroachdb/cockroach-go/v2/crdb/crdbpgx"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/jackc/pgx/v4"
	"github.com/google/uuid"
)

var pool *pgxpool.Pool

type KetoTuple struct {
	ShardID             uuid.UUID
	NetworkID           uuid.UUID
	Namespace           string
	Object              uuid.UUID
	Relation            string
	SubjectID           uuid.UUID
	SubjectSetNamespace *string
	SubjectSetObject    *uuid.UUID
	SubjectSetRelation  *string
	CommitTime          time.Time
}

func Connect(connStr string) error {
	var err error
	pool, err = pgxpool.Connect(context.Background(), connStr)
	if err != nil {
		return fmt.Errorf("unable to connect to CockroachDB: %w", err)
	}
	return nil
}

func Close() {
	if pool != nil {
		pool.Close()
	}
}

func InsertKetoTuple(ctx context.Context, t KetoTuple) error {
	return crdbpgx.ExecuteTx(ctx, pool, pgx.TxOptions{}, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `
			INSERT INTO public.keto_relation_tuples (
				shard_id, nid, namespace, object, relation,
				subject_id, subject_set_namespace,
				subject_set_object, subject_set_relation, commit_time
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		`,
			t.ShardID, t.NetworkID, t.Namespace, t.Object, t.Relation,
			t.SubjectID, t.SubjectSetNamespace,
			t.SubjectSetObject, t.SubjectSetRelation, t.CommitTime,
		)
		return err
	})
}

func InsertUUIDMapping(ctx context.Context, id uuid.UUID, name string) error {
	return crdbpgx.ExecuteTx(ctx, pool, pgx.TxOptions{}, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `
			INSERT INTO public.keto_uuid_mappings (id, string_representation)
			VALUES ($1, $2)
			ON CONFLICT (id) DO NOTHING
		`, id, name)
		return err
	})
}
