package helper

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/mhaatha/HIMA-TI-e-Election/config"
	"github.com/sirupsen/logrus"
)

func CommitOrRollback(ctx context.Context, tx pgx.Tx) {
	if err := tx.Commit(ctx); err != nil {
		if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
			config.Log.WithFields(logrus.Fields{
				"error":         err,
				"rollbackError": rollbackErr,
			})
		} else {
			config.Log.WithError(err).Error("commit failed, rollback succeeded")
		}
	}
}
