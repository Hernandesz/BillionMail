package database_initialization

import (
    "context"
    "github.com/gogf/gf/v2/frame/g"
)

func init() {
    registerHandler(func() {
        statements := []string{
            `CREATE TABLE IF NOT EXISTS bm_master_contacts (
                id BIGSERIAL PRIMARY KEY,
                email VARCHAR(320) NOT NULL UNIQUE,
                full_name TEXT NOT NULL DEFAULT '',
                job_title TEXT NOT NULL DEFAULT '',
                department TEXT NOT NULL DEFAULT '',
                company_name TEXT NOT NULL DEFAULT '',
                industry TEXT NOT NULL DEFAULT '',
                phone TEXT NOT NULL DEFAULT '',
                website TEXT NOT NULL DEFAULT '',
                linkedin_url TEXT NOT NULL DEFAULT '',
                city TEXT NOT NULL DEFAULT '',
                state_region TEXT NOT NULL DEFAULT '',
                country TEXT NOT NULL DEFAULT '',
                verification_status VARCHAR(40) NOT NULL DEFAULT 'unverified',
                consent_status VARCHAR(40) NOT NULL DEFAULT 'not_documented',
                metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
                source_details JSONB NOT NULL DEFAULT '[]'::jsonb,
                created_at BIGINT NOT NULL DEFAULT EXTRACT(EPOCH FROM NOW()),
                updated_at BIGINT NOT NULL DEFAULT EXTRACT(EPOCH FROM NOW()),
                CONSTRAINT bm_master_email_normalized CHECK (email = LOWER(BTRIM(email)))
            )`,
            `CREATE INDEX IF NOT EXISTS idx_bm_master_contacts_company ON bm_master_contacts(company_name)`,
            `CREATE INDEX IF NOT EXISTS idx_bm_master_contacts_updated ON bm_master_contacts(updated_at DESC)`,
        }
        for _, stmt := range statements {
            if _, err := g.DB().Exec(context.Background(),stmt); err != nil {
                g.Log().Error(context.Background(),"initialize central contacts table:",err)
            }
        }
    })
}
