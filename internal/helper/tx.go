package helper

import (
	"database/sql"
	"errors"

	"github.com/mhaatha/HIMA-TI-e-Election/internal/config"
	"github.com/sirupsen/logrus"
)

func RollbackQuietly(tx *sql.Tx) {
	if rollbackErr := tx.Rollback(); rollbackErr != nil {
		if !errors.Is(rollbackErr, sql.ErrTxDone) {
			config.Log.WithFields(logrus.Fields{
				"error": rollbackErr,
			}).Error("failed to rollback transaction")
		}
	}
}
