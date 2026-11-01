package contact

import (
    "billionmail-core/api/contact/v1"
    "billionmail-core/internal/service/public"
    "context"
    "encoding/json"
    "errors"
    "fmt"
    "regexp"
    "strings"

    "github.com/gogf/gf/v2/database/gdb"
    "github.com/gogf/gf/v2/errors/gerror"
    "github.com/gogf/gf/v2/frame/g"
    "github.com/gogf/gf/v2/util/gconv"
)

var masterEmailRE = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]{2,}$`)

func (c *ControllerV1) ImportMasterContacts(ctx context.Context, req *v1.ImportMasterContactsReq) (res *v1.ImportMasterContactsRes, err error) {
    res = &v1.ImportMasterContactsRes{}
    if len(req.Contacts) == 0 || len(req.Contacts) > 500 {
        res.SetError(gerror.New("Import accepts 1 to 500 contacts per batch"))
        return res,nil
    }
    // Validate the complete batch before writing a single record.
    emails := make(map[string]bool,len(req.Contacts))
    for _, contact := range req.Contacts {
        email := strings.ToLower(strings.TrimSpace(contact.Email))
        if len(email) > 320 || !masterEmailRE.MatchString(email) {
            res.SetError(gerror.New("Invalid email in master import"))
            return res,nil
        }
        if emails[email] {
            res.SetError(gerror.New("Duplicated email within import batch: "+email))
            return res,nil
        }
        emails[email] = true
    }

    transactionErr := g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
        for _, item := range req.Contacts {
            email := strings.ToLower(strings.TrimSpace(item.Email))
            meta := item.Metadata
            if meta == nil { meta = map[string]interface{}{} }
            sources := item.SourceDetails
            if sources == nil { sources = []interface{}{} }
            metadataJSON, e := json.Marshal(meta)
            if e != nil { return e }
            sourcesJSON, e := json.Marshal(sources)
            if e != nil { return e }
            if len(metadataJSON)+len(sourcesJSON) > 256000 {
                return errors.New("metadata and sources exceed 256 KB for "+email)
            }

            found, e := tx.GetOne(ctx, "SELECT id FROM bm_master_contacts WHERE email=$1", email)
            if e != nil { return e }
            if _, e = tx.Exec(ctx, `
                INSERT INTO bm_master_contacts
                  (email,full_name,job_title,department,company_name,industry,
                   phone,website,linkedin_url,city,state_region,country,metadata,source_details)
                VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13::jsonb,$14::jsonb)
                ON CONFLICT (email) DO UPDATE SET
                    full_name=EXCLUDED.full_name,
                    job_title=EXCLUDED.job_title,
                    department=EXCLUDED.department,
                    company_name=EXCLUDED.company_name,
                    industry=EXCLUDED.industry,
                    phone=EXCLUDED.phone,
                    website=EXCLUDED.website,
                    linkedin_url=EXCLUDED.linkedin_url,
                    city=EXCLUDED.city,
                    state_region=EXCLUDED.state_region,
                    country=EXCLUDED.country,
                    metadata=EXCLUDED.metadata,
                    source_details=EXCLUDED.source_details,
                    updated_at=EXTRACT(EPOCH FROM NOW())
                `,
                email,strings.TrimSpace(item.FullName),strings.TrimSpace(item.JobTitle),
                strings.TrimSpace(item.Department),strings.TrimSpace(item.CompanyName),
                strings.TrimSpace(item.Industry),strings.TrimSpace(item.Phone),
                strings.TrimSpace(item.Website),strings.TrimSpace(item.LinkedInURL),
                strings.TrimSpace(item.City),strings.TrimSpace(item.StateRegion),
                strings.TrimSpace(item.Country),string(metadataJSON),string(sourcesJSON)); e != nil {
                return fmt.Errorf("save %s: %w",email,e)
            }
            if len(found) > 0 { res.Data.Updated++ } else { res.Data.Added++ }
        }
        return nil
    })
    if transactionErr != nil {
        res.Data.Added, res.Data.Updated = 0,0
        res.SetError(gerror.New("Master import failed: "+transactionErr.Error()))
        return res,nil
    }
    res.Data.Total = res.Data.Added+res.Data.Updated
    _ = gconv.Int(res.Data.Total)
    res.SetSuccess(public.LangCtx(ctx,"Master contacts saved without group assignment or subscription changes"))
    return res,nil
}
