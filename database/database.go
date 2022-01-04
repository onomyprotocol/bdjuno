package database

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/lib/pq"

	"github.com/desmos-labs/juno/db"
	"github.com/desmos-labs/juno/db/postgresql"
	"github.com/desmos-labs/juno/types"
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

// SaveTx implements db.Database
func (db *Db) SaveTx(tx *types.Tx) error {
	sqlStatement := `
INSERT INTO transaction 
    (hash, height, success, messages, memo, signatures, signer_infos, fee, gas_wanted, gas_used, raw_log, logs) 
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12) ON CONFLICT DO NOTHING`

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

	// fix unicode zeros
	escapedLogsBz := strings.ReplaceAll(string(logsBz), `\u0000`, ``)

	_, err = db.Sql.Exec(sqlStatement,
		tx.TxHash, tx.Height, tx.Successful(),
		msgsBz, tx.Body.Memo, pq.Array(sigs),
		sigInfoBz, string(feeBz),
		tx.GasWanted, tx.GasUsed, tx.RawLog, escapedLogsBz,
	)
	return err
}
