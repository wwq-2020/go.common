package gormx

import (
	"github.com/wwq-2020/go.common/errorsx"
	"github.com/wwq-2020/go.common/log"
	"github.com/wwq-2020/go.common/sqlx"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Conf Conf
type Conf struct {
	*sqlx.Conf `toml:",inline" json:",inline" yaml:",inline"`
}

// MustOpen MustOpen
func MustOpen(conf *Conf) *gorm.DB {
	db, err := Open(conf)
	if err != nil {
		log.Fatalf("failed to Open,err:%v", err)
	}
	return db
}

// Open Open
func Open(conf *Conf) (*gorm.DB, error) {
	stdDB, err := sqlx.Open(conf.Conf)
	if err != nil {
		return nil, errorsx.Trace(err)
	}
	mysqlConfig := mysql.Config{
		Conn: stdDB,
	}
	dialector := mysql.New(mysqlConfig)
	gormDB, err := gorm.Open(dialector, &gorm.Config{
		Logger:   logger.Discard,
		ConnPool: stdDB,
	})
	if err != nil {
		return nil, errorsx.Trace(err)
	}
	return gormDB, nil
}
