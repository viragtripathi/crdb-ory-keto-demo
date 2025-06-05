package db

import (
	"context"
	"fmt"
)

func InitSchema() error {
	ctx := context.Background()

	schema := []string{
		`CREATE TABLE IF NOT EXISTS keto_relation_tuples (
			shard_id UUID NOT NULL,
			nid UUID NOT NULL,
			namespace STRING NOT NULL,
			object UUID NOT NULL,
			relation STRING NOT NULL,
			subject_id UUID NULL,
			subject_set_namespace STRING NULL,
			subject_set_object UUID NULL,
			subject_set_relation STRING NULL,
			commit_time TIMESTAMPTZ NOT NULL,
			CONSTRAINT keto_relation_tuples_pkey PRIMARY KEY (shard_id, nid),
			CONSTRAINT chk_keto_rt_subject_type CHECK (
				(
					subject_id IS NULL AND
					subject_set_namespace IS NOT NULL AND
					subject_set_object IS NOT NULL AND
					subject_set_relation IS NOT NULL
				) OR (
					subject_id IS NOT NULL AND
					subject_set_namespace IS NULL AND
					subject_set_object IS NULL AND
					subject_set_relation IS NULL
				)
			)
		)`,
		`CREATE TABLE IF NOT EXISTS keto_uuid_mappings (
			id UUID NOT NULL,
			string_representation STRING NOT NULL,
			CONSTRAINT keto_uuid_mappings_pkey PRIMARY KEY (id)
		)`,
	}

	for _, stmt := range schema {
		if _, err := pool.Exec(ctx, stmt); err != nil {
			return fmt.Errorf("error running schema DDL: %w", err)
		}
	}

	return nil
}
