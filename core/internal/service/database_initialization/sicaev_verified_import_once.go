package database_initialization

import (
    "context"
    "github.com/gogf/gf/v2/database/gdb"
    "github.com/gogf/gf/v2/frame/g"
)

// One-time status reconciliation for the historical SQL import.
// This does NOT create subscribers, groups, campaign assignments or marketing consent.
func init() {
    registerHandler(func() {
        ctx := context.Background()
        err := g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
            if _, err := tx.Exec(`CREATE TABLE IF NOT EXISTS bm_sicaev_migrations (
                name TEXT PRIMARY KEY,
                executed_at BIGINT NOT NULL DEFAULT EXTRACT(EPOCH FROM NOW())
            )`); err != nil {
                return err
            }
            claim, err := tx.Exec(`INSERT INTO bm_sicaev_migrations (name)
                VALUES ('master_sql_verified_20261101')
                ON CONFLICT (name) DO NOTHING`)
            if err != nil {
                return err
            }
            claimed, err := claim.RowsAffected()
            if err != nil {
                return err
            }
            if claimed == 0 {
                return nil
            }
            _, err = tx.Exec(`UPDATE bm_master_contacts
                SET verification_status = 'verified',
                    updated_at = EXTRACT(EPOCH FROM NOW())
                WHERE verification_status <> 'verified'
                  AND metadata @> '{"origin_tables":["companies_contacts","companies_campaign_recipients"]}'::jsonb
                  AND (
                    metadata #>> '{source_contact,verification_status}' = 'verified'
                    OR metadata @> '{"recipient_summary":{"permission_statuses":["verified"]}}'::jsonb
                  )`)
            return err
        })
        if err != nil {
            g.Log().Error(ctx, "one-time central directory source-verification reconciliation: ", err)
        } else {
            g.Log().Info(ctx, "one-time central directory source-verification reconciliation completed")
        }
    })
}
