package activity

import (
	"context"
	"database/sql"
	"encoding/json"
)

type Logger struct {
	db *sql.DB
}

func New(db *sql.DB) *Logger {
	return &Logger{db: db}
}

func (l *Logger) Write(ctx context.Context, category, actorType, actorID, action, targetType, targetID, ip string, detail interface{}) {
	if l == nil || l.db == nil {
		return
	}
	var detailJSON string
	if detail != nil {
		b, _ := json.Marshal(detail)
		detailJSON = string(b)
	} else {
		detailJSON = "{}"
	}
	_, _ = l.db.ExecContext(ctx, `
		INSERT INTO activity_logs (category, actor_type, actor_id, action, target_type, target_id, detail, ip_address)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		category, actorType, actorID, action, targetType, targetID, detailJSON, ip)
}
