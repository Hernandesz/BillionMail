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

            found, e := tx.GetOne("SELECT id FROM bm_master_contacts WHERE email=?", email)
            if e != nil { return e }
            rowJSON, e := json.Marshal(map[string]interface{}{
                "email": email,
                "full_name": strings.TrimSpace(item.FullName),
                "job_title": strings.TrimSpace(item.JobTitle),
                "department": strings.TrimSpace(item.Department),
                "company_name": strings.TrimSpace(item.CompanyName),
                "industry": strings.TrimSpace(item.Industry),
                "phone": strings.TrimSpace(item.Phone),
                "website": strings.TrimSpace(item.Website),
                "linkedin_url": strings.TrimSpace(item.LinkedInURL),
                "city": strings.TrimSpace(item.City),
                "state_region": strings.TrimSpace(item.StateRegion),
                "country": strings.TrimSpace(item.Country),
                "metadata": json.RawMessage(metadataJSON),
                "source_details": json.RawMessage(sourcesJSON),
            })
            if e != nil { return e }
            if _, e = tx.Exec(`
                INSERT INTO bm_master_contacts
                  (email,full_name,job_title,department,company_name,industry,
                   phone,website,linkedin_url,city,state_region,country,metadata,source_details)
                SELECT email, full_name, job_title, department, company_name, industry,
                       phone, website, linkedin_url, city, state_region, country, metadata, source_details
                FROM jsonb_to_record(?::jsonb) AS contact_row (
                    email text, full_name text, job_title text, department text,
                    company_name text, industry text, phone text, website text,
                    linkedin_url text, city text, state_region text, country text,
                    metadata jsonb, source_details jsonb
                )
                ON CONFLICT (email) DO UPDATE SET
                    full_name=COALESCE(NULLIF(EXCLUDED.full_name, ''), bm_master_contacts.full_name),
                    job_title=COALESCE(NULLIF(EXCLUDED.job_title, ''), bm_master_contacts.job_title),
                    department=COALESCE(NULLIF(EXCLUDED.department, ''), bm_master_contacts.department),
                    company_name=COALESCE(NULLIF(EXCLUDED.company_name, ''), bm_master_contacts.company_name),
                    industry=COALESCE(NULLIF(EXCLUDED.industry, ''), bm_master_contacts.industry),
                    phone=COALESCE(NULLIF(EXCLUDED.phone, ''), bm_master_contacts.phone),
                    website=COALESCE(NULLIF(EXCLUDED.website, ''), bm_master_contacts.website),
                    linkedin_url=COALESCE(NULLIF(EXCLUDED.linkedin_url, ''), bm_master_contacts.linkedin_url),
                    city=COALESCE(NULLIF(EXCLUDED.city, ''), bm_master_contacts.city),
                    state_region=COALESCE(NULLIF(EXCLUDED.state_region, ''), bm_master_contacts.state_region),
                    country=COALESCE(NULLIF(EXCLUDED.country, ''), bm_master_contacts.country),
                    metadata=bm_master_contacts.metadata || EXCLUDED.metadata,
                    source_details=(
                        SELECT COALESCE(jsonb_agg(src), '[]'::jsonb)
                        FROM (
                            SELECT DISTINCT value AS src
                            FROM jsonb_array_elements(
                                COALESCE(bm_master_contacts.source_details, '[]'::jsonb) ||
                                COALESCE(EXCLUDED.source_details, '[]'::jsonb)
                            )
                        ) AS distinct_sources
                    ),
                    updated_at=EXTRACT(EPOCH FROM NOW())
                `, string(rowJSON)); e != nil {
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
    res.SetSuccess(public.LangCtx(ctx,"Master contacts saved without group assignment or subscription changes"))
    return res,nil
}
