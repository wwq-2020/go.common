package sqlx

import (
	"context"
	"database/sql"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/wwq-2020/go.common/errorsx"
)

// Conf Conf
type Conf struct {
	User            string `json:"user" toml:"user" yaml:"user"`
	Password        string `json:"password" toml:"password" yaml:"password"`
	Host            string `json:"host" toml:"host" yaml:"host"`
	DBName          string `json:"db_name" toml:"db_name" yaml:"db_name"`
	MaxOpenConns    int    `json:"max_open_conns" toml:"max_open_conns" yaml:"max_open_conns"`
	MaxIdleConns    int    `json:"max_idle_conns" toml:"max_idle_conns" yaml:"max_idle_conns"`
	ConnMaxLifetime int    `json:"conn_max_lifetime" toml:"conn_max_lifetime" yaml:"conn_max_lifetime"`
}

// Fill Fill
func (c *Conf) Fill() {
	if c.MaxOpenConns == 0 {
		c.MaxIdleConns = 10
	}
	if c.MaxIdleConns == 0 {
		c.MaxIdleConns = 10
	}
	if c.ConnMaxLifetime == 0 {
		c.ConnMaxLifetime = 3600
	}
}

// Stmt Stmt
type Stmt interface {
	PrepareContext(ctx context.Context, query string) (PreparedStmt, error)
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) Row
}

// DB DB
type DB interface {
	Stmt
	BeginTx(ctx context.Context, opts *sql.TxOptions) (Tx, error)
	Close() error
	Conn(ctx context.Context) (Conn, error)
	PingContext(ctx context.Context) error
}

type db struct {
	db *sql.DB
}

// Open Open
func Open(conf *Conf) (DB, error) {
	conf.Fill()
	mysqlConfig := mysql.Config{
		User:   conf.User,
		Passwd: conf.Password,
		Net:    "tcp",
		Addr:   conf.Host,
		DBName: conf.DBName,
		Params: map[string]string{
			"time_zone": "'Asia/Shanghai'",
		},
		Collation:            "utf8mb4_bin",
		Loc:                  time.FixedZone("Asia/Shanghai", 8*60*60),
		Timeout:              1000 * time.Millisecond,
		AllowNativePasswords: true,
		CheckConnLiveness:    true,
		InterpolateParams:    true,
		ParseTime:            true,
	}
	stdDB, err := sql.Open("mysql", mysqlConfig.FormatDSN())
	if err != nil {
		return nil, errorsx.Trace(err)
	}
	stdDB.SetMaxOpenConns(conf.MaxOpenConns)
	stdDB.SetMaxIdleConns(conf.MaxIdleConns)
	stdDB.SetConnMaxLifetime(time.Duration(conf.ConnMaxLifetime) * time.Second)
	return &db{db: stdDB}, nil
}

func (db *db) BeginTx(ctx context.Context, opts *sql.TxOptions) (Tx, error) {
	stdTx, err := db.db.BeginTx(ctx, opts)
	if err != nil {
		return nil, errorsx.Trace(err)
	}
	return &tx{Tx: stdTx}, nil
}

func (db *db) Close() error {
	if err := db.db.Close(); err != nil {
		return errorsx.Trace(err)
	}
	return nil
}

func (db *db) Conn(ctx context.Context) (Conn, error) {
	stdConn, err := db.db.Conn(ctx)
	if err != nil {
		return nil, errorsx.Trace(err)
	}
	return &conn{Conn: stdConn}, nil
}

func (db *db) PingContext(ctx context.Context) error {
	if err := db.db.PingContext(ctx); err != nil {
		return errorsx.Trace(err)
	}
	return nil
}

func (db *db) PrepareContext(ctx context.Context, query string) (PreparedStmt, error) {
	stdStmt, err := db.db.PrepareContext(ctx, query)
	if err != nil {
		return nil, errorsx.Trace(err)
	}
	return &stmt{Stmt: stdStmt}, nil
}

func (db *db) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	var cancel context.CancelFunc
	if _, ok := ctx.Deadline(); !ok {
		ctx, cancel = context.WithTimeout(ctx, time.Second)
		defer func() {
			cancel()
		}()
	}
	result, err := db.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, errorsx.Trace(err)
	}
	return result, nil
}

func (db *db) QueryContext(ctx context.Context, query string, args ...interface{}) (r Rows, err error) {
	var cancel context.CancelFunc
	if _, ok := ctx.Deadline(); !ok {
		ctx, cancel = context.WithTimeout(ctx, time.Second)
		defer func() {
			if err != nil {
				cancel()
			}
		}()
	}
	stdRows, err := db.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, errorsx.Trace(err)
	}
	return &rows{Rows: stdRows, cancel: cancel}, nil
}

func (db *db) QueryRowContext(ctx context.Context, query string, args ...interface{}) Row {
	var cancel context.CancelFunc
	if _, ok := ctx.Deadline(); !ok {
		ctx, cancel = context.WithTimeout(ctx, time.Second)
	}
	stdRow := db.db.QueryRowContext(ctx, query, args...)
	return &row{Row: stdRow, cancel: cancel}
}
