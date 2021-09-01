package sqlx

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/wwq-2020/go.common/errorsx"
	"github.com/wwq-2020/go.common/log"
)

// Conf Conf
type Conf struct {
	User            string `json:"user" toml:"user" yaml:"user"`
	Password        string `json:"password" toml:"password" yaml:"password"`
	Host            string `json:"host" toml:"host" yaml:"host"`
	Port            int32  `json:"port" toml:"port" yaml:"port"`
	DBName          string `json:"db_name" toml:"db_name" yaml:"db_name"`
	Charset         string `json:"charset" toml:"charset" yaml:"charset"`
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

// MustOpen MustOpen
func MustOpen(conf *Conf) *sql.DB {
	stdDB, err := Open(conf)
	if err != nil {
		log.WithError(err).
			Fatal("failed to Open")
	}
	return stdDB
}

// Open Open
func Open(conf *Conf) (*sql.DB, error) {
	conf.Fill()
	addr := fmt.Sprintf("%s:%d", conf.Host, conf.Port)
	mysqlConfig := mysql.Config{
		User:   conf.User,
		Passwd: conf.Password,
		Net:    "tcp",
		Addr:   addr,
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
	dsn := mysqlConfig.FormatDSN()
	connector := NewMysqlConnector(dsn)
	stdDB := sql.OpenDB(connector)
	stdDB.SetMaxOpenConns(conf.MaxOpenConns)
	stdDB.SetMaxIdleConns(conf.MaxIdleConns)
	stdDB.SetConnMaxLifetime(time.Duration(conf.ConnMaxLifetime) * time.Second)
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()
	if err := stdDB.PingContext(ctx); err != nil {
		return nil, errorsx.Trace(err)
	}
	return stdDB, nil
}

// MustOpenStd MustOpenStd
func MustOpenStd(conf *Conf) *sql.DB {
	stdDB, err := OpenStd(conf)
	if err != nil {
		log.WithError(err).
			Fatal("failed to OpenStd")
	}
	return stdDB
}

// OpenStd OpenStd
func OpenStd(conf *Conf) (*sql.DB, error) {
	conf.Fill()
	addr := fmt.Sprintf("%s:%d", conf.Host, conf.Port)
	mysqlConfig := mysql.Config{
		User:   conf.User,
		Passwd: conf.Password,
		Net:    "tcp",
		Addr:   addr,
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
	dsn := mysqlConfig.FormatDSN()
	connector := NewMysqlStdConnector(dsn)
	stdDB := sql.OpenDB(connector)
	stdDB.SetMaxOpenConns(conf.MaxOpenConns)
	stdDB.SetMaxIdleConns(conf.MaxIdleConns)
	stdDB.SetConnMaxLifetime(time.Duration(conf.ConnMaxLifetime) * time.Second)
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()
	if err := stdDB.PingContext(ctx); err != nil {
		return nil, errorsx.Trace(err)
	}
	return stdDB, nil
}
