package sqlx

import (
	"context"
	"database/sql/driver"

	"github.com/go-sql-driver/mysql"
	"github.com/wwq-2020/go.common/errorsx"
)

type dsnConnector struct {
	dsn    string
	driver driver.Driver
}

func (t dsnConnector) Connect(_ context.Context) (driver.Conn, error) {
	conn, err := t.driver.Open(t.dsn)
	if err != nil {
		return nil, errorsx.Trace(err)
	}
	return conn, nil
}

func (t dsnConnector) Driver() driver.Driver {
	return t.driver
}

// NewMysqlConnector NewMysqlConnector
func NewMysqlConnector(dsn string) driver.Connector {
	return NewConnector(dsn, &mysql.MySQLDriver{})
}

// NewConnector NewConnector
func NewConnector(dsn string, driver driver.Driver) driver.Connector {
	return &dsnConnector{
		dsn:    dsn,
		driver: WrapDriver(driver),
	}
}

// NewMysqlStdConnector NewMysqlStdConnector
func NewMysqlStdConnector(dsn string) driver.Connector {
	return NewStdConnector(dsn, &mysql.MySQLDriver{})
}

// NewStdConnector NewStdConnector
func NewStdConnector(dsn string, driver driver.Driver) driver.Connector {
	return &dsnConnector{
		dsn:    dsn,
		driver: driver,
	}
}
