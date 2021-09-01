package gormx

import (
	"github.com/wwq-2020/go.common/log"
	"github.com/wwq-2020/go.common/tracing"
	"gorm.io/gorm"
)

// TracingPlugin TracingPlugin
type TracingPlugin struct{}

var _ gorm.Plugin = &TracingPlugin{}

// Name Name
func (p *TracingPlugin) Name() string {
	return "tracing"
}

const (
	tracingBefore = "tracing:before"
	tracingAfter  = "tracing:after"
)

// Initialize InitializeInitialize
func (p *TracingPlugin) Initialize(db *gorm.DB) error {
	callback := "create"
	db.Callback().Create().Before(callback).Register(tracingBefore, traceStart(callback))
	db.Callback().Create().After(callback).Register(tracingAfter, traceEnd(callback))

	callback = "Delete"
	db.Callback().Delete().Before(callback).Register(tracingBefore, traceStart(callback))
	db.Callback().Delete().After(callback).Register(tracingAfter, traceEnd(callback))

	callback = "Update"
	db.Callback().Update().Before(callback).Register(tracingBefore, traceStart(callback))
	db.Callback().Update().After(callback).Register(tracingAfter, traceEnd(callback))

	callback = "Query"
	db.Callback().Query().Before(callback).Register(tracingBefore, traceStart(callback))
	db.Callback().Query().After(callback).Register(tracingAfter, traceEnd(callback))

	callback = "Raw"
	db.Callback().Raw().Before(callback).Register(tracingBefore, traceStart(callback))
	db.Callback().Raw().After(callback).Register(tracingAfter, traceEnd(callback))

	callback = "Row"
	db.Callback().Row().Before(callback).Register(tracingBefore, traceStart(callback))
	db.Callback().Row().After(callback).Register(tracingAfter, traceEnd(callback))
	return nil

}

func traceStart(op string) func(db *gorm.DB) {
	return func(db *gorm.DB) {
		stmt := db.Statement
		if stmt == nil {
			return
		}
		ctx := stmt.Context
		_, ctx = tracing.StartSpan(ctx, "db")
		stmt.Context = ctx
	}
}

func traceEnd(op string) func(db *gorm.DB) {
	return func(db *gorm.DB) {
		stmt := db.Statement
		if stmt == nil {
			return
		}

		ctx := stmt.Context
		span := tracing.SpanFromContext(ctx)
		span.Finish(&db.Error)
		log.WithField("vars", stmt.Vars).
			WithField("sql", stmt.SQL.String()).
			InfoContext(ctx, "tracing")
	}
}
