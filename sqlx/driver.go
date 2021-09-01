package sqlx

import (
	"database/sql/driver"

	"github.com/go-sql-driver/mysql"
	"github.com/wwq-2020/go.common/errorsx"
)

type wrappedDriver struct {
	driver driver.Driver
}

func (d *wrappedDriver) Open(name string) (driver.Conn, error) {
	conn, err := d.driver.Open(name)
	if err != nil {
		return nil, errorsx.Trace(err)
	}
	return NewConn(conn), nil
}

// WrappedMysqlDriver WrappedMysqlDriver
func WrappedMysqlDriver() driver.Driver {
	return WrapDriver(&mysql.MySQLDriver{})
}

// WrapDriver WrapDriver
func WrapDriver(driver driver.Driver) driver.Driver {
	return &wrappedDriver{
		driver: driver,
	}
}
