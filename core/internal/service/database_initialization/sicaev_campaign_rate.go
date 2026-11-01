package database_initialization

import (
    "context"
    "github.com/gogf/gf/v2/frame/g"
)

// Persisted, shared campaign dispatch rate limits. Does not change SMTP or API mail.
// A single PostgreSQL lock coordinates all campaign executors and survives restarts.
func init() {
    registerHandler(func() {
        ctx := context.Background()
        sql := []string{
            `ALTER TABLE email_tasks ADD COLUMN IF NOT EXISTS max_per_hour INTEGER NOT NULL DEFAULT 20`,
            `ALTER TABLE email_tasks ADD COLUMN IF NOT EXISTS max_per_day INTEGER NOT NULL DEFAULT 200`,
            `CREATE TABLE IF NOT EXISTS bm_sicaev_campaign_rate (
                scope VARCHAR(64) PRIMARY KEY,
                last_at_ms BIGINT NOT NULL DEFAULT 0,
                day_key INTEGER NOT NULL DEFAULT 0,
                day_count INTEGER NOT NULL DEFAULT 0
            )`,
            `INSERT INTO bm_sicaev_campaign_rate (scope)
                VALUES ('global') ON CONFLICT (scope) DO NOTHING`,
        }
        for _, statement := range sql {
            if _, err := g.DB().Exec(ctx,statement); err != nil {
                g.Log().Error(ctx,"SICAEV campaign rate schema initialization failed: ",err)
            }
        }
    })
}
