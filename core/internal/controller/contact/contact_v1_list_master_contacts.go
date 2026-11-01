package contact

import (
    "billionmail-core/api/contact/v1"
    "billionmail-core/internal/service/public"
    "context"
    "encoding/json"
    "strings"

    "github.com/gogf/gf/v2/errors/gerror"
    "github.com/gogf/gf/v2/frame/g"
    "github.com/gogf/gf/v2/util/gconv"
)

const masterContactFrom = `
FROM (
    SELECT email FROM bm_master_contacts
    UNION
    SELECT LOWER(BTRIM(email)) FROM bm_contacts
) directory
LEFT JOIN bm_master_contacts m ON m.email = directory.email
LEFT JOIN LATERAL (
    SELECT
        COUNT(*) AS native_count,
        COALESCE(MAX(CASE WHEN c.active = 1 AND c.status = 1 THEN 2
                          WHEN c.active = 1 THEN 1 ELSE 0 END), 0) AS delivery_state,
        COALESCE(jsonb_agg(DISTINCT jsonb_build_object('id', grp.id, 'name', grp.name))
                 FILTER (WHERE grp.id IS NOT NULL), '[]'::jsonb) AS groups,
        COALESCE(jsonb_agg(DISTINCT jsonb_build_object('id', tag.id, 'name', tag.name))
                 FILTER (WHERE tag.id IS NOT NULL), '[]'::jsonb) AS tags
    FROM bm_contacts c
    LEFT JOIN bm_contact_groups grp ON c.group_id = grp.id
    LEFT JOIN bm_contact_tags ct ON ct.contact_id = c.id
    LEFT JOIN bm_tags tag ON tag.id = ct.tag_id
    WHERE LOWER(BTRIM(c.email)) = directory.email
) nc ON TRUE
`

func masterWhere(filter string) string {
    condition := ` WHERE (directory.email ILIKE $1 OR
        COALESCE(m.company_name,'') ILIKE $1 OR
        COALESCE(m.full_name,'') ILIKE $1 OR
        COALESCE(m.job_title,'') ILIKE $1 OR
        COALESCE(m.department,'') ILIKE $1) `
    switch filter {
    case "ungrouped":
        condition += " AND nc.groups = '[]'::jsonb"
    case "grouped":
        condition += " AND nc.groups <> '[]'::jsonb"
    case "master":
        condition += " AND m.id IS NOT NULL"
    case "native":
        condition += " AND nc.native_count > 0"
    case "unverified":
        condition += " AND COALESCE(m.verification_status,'unverified') <> 'verified'"
    case "subscribed":
        condition += " AND nc.delivery_state = 2"
    }
    return condition
}

func parseMasterJSON(raw string, dest interface{}) {
    if raw != "" {
        _ = json.Unmarshal([]byte(raw),dest)
    }
}

func (c *ControllerV1) ListMasterContacts(ctx context.Context, req *v1.ListMasterContactsReq) (res *v1.ListMasterContactsRes, err error) {
    res = &v1.ListMasterContactsRes{}
    page, pageSize := req.Page, req.PageSize
    if page < 1 { page = 1 }
    if pageSize < 1 { pageSize = 20 }
    if pageSize > 100 { pageSize = 100 }
    search := "%" + strings.TrimSpace(req.Keyword) + "%"
    clause := masterWhere(req.Filter)

    countRows, queryErr := g.DB().GetAll(ctx, "SELECT COUNT(*) AS n "+masterContactFrom+clause, search)
    if queryErr != nil {
        res.SetError(gerror.New("Could not retrieve central directory count: "+queryErr.Error()))
        return res,nil
    }
    if len(countRows) > 0 { res.Data.Total = gconv.Int64(countRows[0]["n"]) }

    query := `SELECT COALESCE(m.id, 0) AS id, directory.email,
      COALESCE(m.full_name,'') AS full_name, COALESCE(m.job_title,'') AS job_title,
      COALESCE(m.department,'') AS department, COALESCE(m.company_name,'') AS company_name,
      COALESCE(m.industry,'') AS industry, COALESCE(m.phone,'') AS phone,
      COALESCE(m.website,'') AS website, COALESCE(m.linkedin_url,'') AS linkedin_url,
      COALESCE(m.city,'') AS city, COALESCE(m.state_region,'') AS state_region,
      COALESCE(m.country,'') AS country,
      COALESCE(m.verification_status,'unverified') AS verification_status,
      COALESCE(m.consent_status,'not_documented') AS consent_status,
      CASE WHEN nc.delivery_state=2 THEN 'subscribed'
           WHEN nc.delivery_state=1 THEN 'unconfirmed'
           ELSE 'not_subscribed' END AS native_status,
      CASE WHEN m.id IS NOT NULL AND nc.native_count > 0 THEN 'both'
           WHEN m.id IS NOT NULL THEN 'master' ELSE 'native' END AS source,
      COALESCE(m.updated_at,0) AS updated_at,
      nc.groups::text AS groups, nc.tags::text AS tags,
      COALESCE(m.metadata::text,'{}') AS metadata,
      COALESCE(m.source_details::text,'[]') AS source_details
    ` + masterContactFrom + clause +
    " ORDER BY COALESCE(m.company_name,''), directory.email LIMIT $2 OFFSET $3"

    rows, queryErr := g.DB().GetAll(ctx, query, search, pageSize, (page-1)*pageSize)
    if queryErr != nil {
        res.SetError(gerror.New("Could not retrieve central directory: "+queryErr.Error()))
        return res,nil
    }
    res.Data.List = make([]*v1.MasterContactInfo,0,len(rows))
    for _, row := range rows {
        item := &v1.MasterContactInfo{
            Id: gconv.Int64(row["id"]),
            Email: gconv.String(row["email"]),
            FullName: gconv.String(row["full_name"]),
            JobTitle: gconv.String(row["job_title"]),
            Department: gconv.String(row["department"]),
            CompanyName: gconv.String(row["company_name"]),
            Industry: gconv.String(row["industry"]),
            Phone: gconv.String(row["phone"]),
            Website: gconv.String(row["website"]),
            LinkedInURL: gconv.String(row["linkedin_url"]),
            City: gconv.String(row["city"]),
            StateRegion: gconv.String(row["state_region"]),
            Country: gconv.String(row["country"]),
            VerificationStatus: gconv.String(row["verification_status"]),
            ConsentStatus: gconv.String(row["consent_status"]),
            NativeStatus: gconv.String(row["native_status"]),
            Source: gconv.String(row["source"]),
            UpdatedAt: gconv.Int64(row["updated_at"]),
            Groups: []v1.MasterGroupInfo{},
            Tags: []v1.MasterTagInfo{},
            Metadata: map[string]interface{}{},
            SourceDetails: []interface{}{},
        }
        parseMasterJSON(gconv.String(row["groups"]), &item.Groups)
        parseMasterJSON(gconv.String(row["tags"]), &item.Tags)
        parseMasterJSON(gconv.String(row["metadata"]), &item.Metadata)
        parseMasterJSON(gconv.String(row["source_details"]), &item.SourceDetails)
        res.Data.List = append(res.Data.List,item)
    }
    res.SetSuccess(public.LangCtx(ctx,"Central directory retrieved successfully"))
    return res,nil
}
