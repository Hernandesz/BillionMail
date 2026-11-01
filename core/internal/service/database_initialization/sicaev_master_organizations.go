package database_initialization

import (
    "context"
    "github.com/gogf/gf/v2/frame/g"
)

func init() {
    registerHandler(func() {
        ctx := context.Background()
        stmts := []string{
            `CREATE TABLE IF NOT EXISTS bm_master_organizations (
                id BIGSERIAL PRIMARY KEY,
                organization_key VARCHAR(400) NOT NULL UNIQUE,
                name TEXT NOT NULL,
                domain TEXT NOT NULL DEFAULT '',
                website TEXT NOT NULL DEFAULT '',
                industry TEXT NOT NULL DEFAULT '',
                country TEXT NOT NULL DEFAULT '',
                state_region TEXT NOT NULL DEFAULT '',
                city TEXT NOT NULL DEFAULT '',
                phone TEXT NOT NULL DEFAULT '',
                description TEXT NOT NULL DEFAULT '',
                metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
                source_details JSONB NOT NULL DEFAULT '[]'::jsonb,
                created_at BIGINT NOT NULL DEFAULT EXTRACT(EPOCH FROM NOW()),
                updated_at BIGINT NOT NULL DEFAULT EXTRACT(EPOCH FROM NOW())
            )`,
            `CREATE INDEX IF NOT EXISTS idx_bm_master_org_name ON bm_master_organizations(name)`,
            `CREATE INDEX IF NOT EXISTS idx_bm_master_org_industry ON bm_master_organizations(industry)`,
        }
        for _, stmt := range stmts {
            if _, err := g.DB().Exec(ctx,stmt);err != nil {
                g.Log().Error(ctx,"initialize central organizations table: ",err)
            }
        }
    })
}
