package database

import (
	"encoding/base64"
	"fmt"
	"github.com/forbole/juno/v3/types/config"
	"github.com/lib/pq"
	"strings"

	db "github.com/forbole/juno/v3/database"
	"github.com/forbole/juno/v3/database/postgresql"
	"github.com/forbole/juno/v3/types"
	"github.com/jmoiron/sqlx"
)

var _ db.Database = &Db{}

// Db represents a PostgreSQL database with expanded features.
// so that it can properly store custom BigDipper-related data.
type Db struct {
	*postgresql.Database
	Sqlx *sqlx.DB
}

// Builder allows to create a new Db instance implementing the db.Builder type
func Builder(ctx *db.Context) (db.Database, error) {
	database, err := postgresql.Builder(ctx)
	if err != nil {
		return nil, err
	}

	psqlDb, ok := (database).(*postgresql.Database)
	if !ok {
		return nil, fmt.Errorf("invalid configuration database, must be PostgreSQL")
	}

	return &Db{
		Database: psqlDb,
		Sqlx:     sqlx.NewDb(psqlDb.Sql, "postgresql"),
	}, nil
}

// Cast allows to cast the given db to a Db instance
func Cast(db db.Database) *Db {
	bdDatabase, ok := db.(*Db)
	if !ok {
		panic(fmt.Errorf("given database instance is not a Db"))
	}
	return bdDatabase
}

// SaveTx implements database.Database
func (db *Db) SaveTx(tx *types.Tx) error {
	var partitionID int64

	partitionSize := config.Cfg.Database.PartitionSize
	if partitionSize > 0 {
		partitionID = tx.Height / partitionSize
		err := db.createPartitionIfNotExists("transaction", partitionID)
		if err != nil {
			return err
		}
	}

	return db.saveTxInsidePartition(tx, partitionID)
}

// createPartitionIfNotExists creates a new partition having the given partition id if not existing
func (db *Db) createPartitionIfNotExists(table string, partitionID int64) error {
	partitionTable := fmt.Sprintf("%s_%d", table, partitionID)

	stmt := fmt.Sprintf(
		"CREATE TABLE IF NOT EXISTS %s PARTITION OF %s FOR VALUES IN (%d)",
		partitionTable,
		table,
		partitionID,
	)
	_, err := db.Sql.Exec(stmt)

	if err != nil {
		return err
	}

	return nil
}

// saveTxInsidePartition stores the given transaction inside the partition having the given id
func (db *Db) saveTxInsidePartition(tx *types.Tx, partitionId int64) error {
	sqlStatement := `
INSERT INTO transaction 
(hash, height, success, messages, memo, signatures, signer_infos, fee, gas_wanted, gas_used, raw_log, logs, partition_id) 
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13) ON CONFLICT DO NOTHING`

	var sigs = make([]string, len(tx.Signatures))
	for index, sig := range tx.Signatures {
		sigs[index] = base64.StdEncoding.EncodeToString(sig)
	}

	var msgs = make([]string, len(tx.Body.Messages))
	for index, msg := range tx.Body.Messages {
		bz, err := db.EncodingConfig.Marshaler.MarshalJSON(msg)
		if err != nil {
			return err
		}
		msgs[index] = string(bz)
	}
	msgsBz := fmt.Sprintf("[%s]", strings.Join(msgs, ","))

	feeBz, err := db.EncodingConfig.Marshaler.MarshalJSON(tx.AuthInfo.Fee)
	if err != nil {
		return fmt.Errorf("failed to JSON encode tx fee: %s", err)
	}

	var sigInfos = make([]string, len(tx.AuthInfo.SignerInfos))
	for index, info := range tx.AuthInfo.SignerInfos {
		bz, err := db.EncodingConfig.Marshaler.MarshalJSON(info)
		if err != nil {
			return err
		}
		sigInfos[index] = string(bz)
	}
	sigInfoBz := fmt.Sprintf("[%s]", strings.Join(sigInfos, ","))

	logsBz, err := db.EncodingConfig.Amino.MarshalJSON(tx.Logs)
	if err != nil {
		return err
	}
	// this line is and only change required for the fix for gravity logs
	logsBzString := strings.ReplaceAll(string(logsBz), "\\u0000", "")

	_, err = db.Sql.Exec(sqlStatement,
		tx.TxHash, tx.Height, tx.Successful(),
		msgsBz, tx.Body.Memo, pq.Array(sigs),
		sigInfoBz, string(feeBz),
		tx.GasWanted, tx.GasUsed, tx.RawLog, logsBzString,
		partitionId,
	)
	return err
}
