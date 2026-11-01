package database_initialization

import (
    "context"
    "fmt"

    "github.com/gogf/gf/v2/database/gdb"
    "github.com/gogf/gf/v2/frame/g"
    "github.com/gogf/gf/v2/util/gconv"
)

// One-time owner-requested activation of existing memberships for ISO course groups.
// Does not create contacts, opt-in evidence, events, memberships, or email tasks.
// Preserve a database audit of the previous status and active values.
func init() {
    registerHandler(func() {
        ctx := context.Background()
        const migration = "sicaev_iso_groups_7_8_activate_20261101"
        err := g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
            // Lock this migration even if two application instances start simultaneously.
            if _, err := tx.Exec("SELECT pg_advisory_xact_lock(20261101, 78)"); err != nil {
                return err
            }
            if _, err := tx.Exec(`CREATE TABLE IF NOT EXISTS bm_sicaev_migrations (
                name TEXT PRIMARY KEY,
                executed_at BIGINT NOT NULL DEFAULT EXTRACT(EPOCH FROM NOW())
            )`); err != nil {
                return err
            }
            already, err := tx.GetOne("SELECT name FROM bm_sicaev_migrations WHERE name = '"+migration+"'")
            if err != nil {
                return err
            }
            if already != nil {
                return nil
            }

            counts, err := tx.GetOne(`SELECT
                COUNT(*) AS memberships,
                COUNT(DISTINCT LOWER(BTRIM(c.email))) AS unique_emails,
                COUNT(*) FILTER (WHERE c.group_id = 7) AS group_7,
                COUNT(*) FILTER (WHERE c.group_id = 8) AS group_8,
                COUNT(*) FILTER (WHERE m.verification_status = 'verified') AS verified_memberships
                FROM bm_contacts c
                LEFT JOIN bm_master_contacts m ON m.email = LOWER(BTRIM(c.email))
                WHERE c.group_id IN (7,8)`)
            if err != nil {
                return err
            }
            if counts == nil ||
                gconv.Int(counts["memberships"]) != 178 ||
                gconv.Int(counts["unique_emails"]) != 89 ||
                gconv.Int(counts["group_7"]) != 89 ||
                gconv.Int(counts["group_8"]) != 89 ||
                gconv.Int(counts["verified_memberships"]) != 178 {
                return fmt.Errorf("ISO groups precondition not met; refusing status update: %v", counts)
            }

            if _, err := tx.Exec(`CREATE TABLE IF NOT EXISTS bm_sicaev_membership_audit (
                migration_name TEXT NOT NULL,
                contact_id INTEGER NOT NULL,
                email VARCHAR(320) NOT NULL,
                group_id INTEGER NOT NULL,
                previous_status INTEGER,
                previous_active INTEGER,
                changed_at BIGINT NOT NULL DEFAULT EXTRACT(EPOCH FROM NOW()),
                PRIMARY KEY (migration_name, contact_id)
            )`); err != nil {
                return err
            }
            backup, err := tx.Exec(`INSERT INTO bm_sicaev_membership_audit
                (migration_name, contact_id, email, group_id, previous_status, previous_active)
                SELECT 'sicaev_iso_groups_7_8_activate_20261101',
                       c.id, c.email, c.group_id, c.status, c.active
                FROM bm_contacts c
                INNER JOIN bm_master_contacts m ON m.email = LOWER(BTRIM(c.email))
                WHERE c.group_id IN (7,8) AND m.verification_status = 'verified'
                ON CONFLICT (migration_name, contact_id) DO NOTHING`)
            if err != nil {
                return err
            }
            saved, err := backup.RowsAffected()
            if err != nil {
                return err
            }
            if saved != 178 {
                return fmt.Errorf("expected to back up 178 memberships, got %d", saved)
            }

            change, err := tx.Exec(`UPDATE bm_contacts AS c SET status = 1, active = 1
                FROM bm_master_contacts AS m
                WHERE c.group_id IN (7,8)
                  AND m.email = LOWER(BTRIM(c.email))
                  AND m.verification_status = 'verified'`)
            if err != nil {
                return err
            }
            updated, err := change.RowsAffected()
            if err != nil {
                return err
            }
            if updated != 178 {
                return fmt.Errorf("expected to update 178 memberships, got %d", updated)
            }

            after, err := tx.GetOne(`SELECT COUNT(*) AS subscribed
                FROM bm_contacts WHERE group_id IN (7,8) AND active=1 AND status=1`)
            if err != nil {
                return err
            }
            if after == nil || gconv.Int(after["subscribed"]) != 178 {
                return fmt.Errorf("post-update check failed: %v", after)
            }

            _, err = tx.Exec(`INSERT INTO bm_sicaev_migrations(name)
                VALUES ('sicaev_iso_groups_7_8_activate_20261101')`)
            return err
        })
        if err != nil {
            g.Log().Error(ctx, "SICAEV ISO memberships one-time update ROLLED BACK: ", err)
            return
        }
        g.Log().Info(ctx, "SICAEV ISO memberships one-time migration completed or already applied")
    })
}
