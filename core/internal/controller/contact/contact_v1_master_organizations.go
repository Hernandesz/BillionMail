package contact

import (
    "billionmail-core/api/contact/v1"
    "billionmail-core/internal/service/public"
    "context"
    "encoding/json"
    "fmt"
    "strings"

    "github.com/gogf/gf/v2/database/gdb"
    "github.com/gogf/gf/v2/errors/gerror"
    "github.com/gogf/gf/v2/frame/g"
    "github.com/gogf/gf/v2/util/gconv"
)

func masterOrganizationKey(item v1.MasterOrganizationInput) string {
    domain := strings.ToLower(strings.TrimSpace(item.Domain))
    domain = strings.TrimPrefix(domain, "www.")
    if domain != "" {
        return "domain:" + domain
    }
    name := strings.ToLower(strings.TrimSpace(item.Name))
    location := strings.ToLower(strings.TrimSpace(item.City+"|"+item.StateRegion+"|"+item.Country))
    return "name:" + name + "|" + location
}

func (c *ControllerV1) ImportMasterOrganizations(ctx context.Context, req *v1.ImportMasterOrganizationsReq) (res *v1.ImportMasterOrganizationsRes, err error) {
    res = &v1.ImportMasterOrganizationsRes{}
    if len(req.Organizations) > 100 {
        res.SetError(gerror.New("Maximum 100 organizations per request"))
        return res, nil
    }
    if len(req.Organizations) == 0 {
        res.SetSuccess("No organizations to import")
        return res, nil
    }
    keys := make(map[string]bool, len(req.Organizations))
    for _, org := range req.Organizations {
        if strings.TrimSpace(org.Name) == "" {
            res.SetError(gerror.New("Organization name required"))
            return res, nil
        }
        key := masterOrganizationKey(org)
        if len(key) > 400 || keys[key] {
            res.SetError(gerror.New("Invalid or duplicate organization key in batch"))
            return res, nil
        }
        keys[key] = true
        if strings.Contains(strings.ToLower(org.Name+" "+org.Domain), "fusoni") {
            res.SetError(gerror.New("Excluded organization domain/name"))
            return res, nil
        }
    }

    dbErr := g.DB().Transaction(ctx,func(ctx context.Context, tx gdb.TX) error {
        for _, org := range req.Organizations {
            key := masterOrganizationKey(org)
            found, e := tx.GetOne("SELECT id FROM bm_master_organizations WHERE organization_key=?", key)
            if e != nil { return e }
            meta:=org.Metadata
            if meta==nil {meta=map[string]interface{}{}}
            sources:=org.SourceDetails
            if sources==nil {sources=[]interface{}{}}
            raw, e:=json.Marshal(map[string]interface{}{
                "organization_key":key,
                "name":strings.TrimSpace(org.Name),
                "domain":strings.ToLower(strings.TrimSpace(org.Domain)),
                "website":strings.TrimSpace(org.Website),
                "industry":strings.TrimSpace(org.Industry),
                "country":strings.TrimSpace(org.Country),
                "state_region":strings.TrimSpace(org.StateRegion),
                "city":strings.TrimSpace(org.City),
                "phone":strings.TrimSpace(org.Phone),
                "description":strings.TrimSpace(org.Description),
                "metadata":meta,
                "source_details":sources,
            })
            if e!=nil{return e}
            if len(raw)>256000{return fmt.Errorf("organization record exceeds 256 KB: %s",key)}
            _,e=tx.Exec(`
                INSERT INTO bm_master_organizations
                (organization_key,name,domain,website,industry,country,state_region,
                 city,phone,description,metadata,source_details)
                SELECT organization_key,name,domain,website,industry,country,state_region,
                       city,phone,description,metadata,source_details
                FROM jsonb_to_record(?::jsonb) AS org_row(
                    organization_key text,name text,domain text,website text,industry text,
                    country text,state_region text,city text,phone text,description text,
                    metadata jsonb,source_details jsonb
                )
                ON CONFLICT (organization_key) DO UPDATE SET
                    name=COALESCE(NULLIF(EXCLUDED.name,''),bm_master_organizations.name),
                    domain=COALESCE(NULLIF(EXCLUDED.domain,''),bm_master_organizations.domain),
                    website=COALESCE(NULLIF(EXCLUDED.website,''),bm_master_organizations.website),
                    industry=COALESCE(NULLIF(EXCLUDED.industry,''),bm_master_organizations.industry),
                    country=COALESCE(NULLIF(EXCLUDED.country,''),bm_master_organizations.country),
                    state_region=COALESCE(NULLIF(EXCLUDED.state_region,''),bm_master_organizations.state_region),
                    city=COALESCE(NULLIF(EXCLUDED.city,''),bm_master_organizations.city),
                    phone=COALESCE(NULLIF(EXCLUDED.phone,''),bm_master_organizations.phone),
                    description=COALESCE(NULLIF(EXCLUDED.description,''),bm_master_organizations.description),
                    metadata=bm_master_organizations.metadata||EXCLUDED.metadata,
                    source_details=(
                        SELECT COALESCE(jsonb_agg(source_item),'[]'::jsonb)
                        FROM (
                            SELECT DISTINCT value AS source_item
                            FROM jsonb_array_elements(
                              COALESCE(bm_master_organizations.source_details,'[]'::jsonb) ||
                              COALESCE(EXCLUDED.source_details,'[]'::jsonb)
                            )
                        ) distinct_sources
                    ),
                    updated_at=EXTRACT(EPOCH FROM NOW())
            `,string(raw))
            if e!=nil{return fmt.Errorf("save organization %s: %w",key,e)}
            if found==nil {res.Data.Added++} else {res.Data.Updated++}
        }
        return nil
    })
    if dbErr != nil{
        res.Data.Added,res.Data.Updated=0,0
        res.SetError(gerror.New("Central company import failed: "+dbErr.Error()))
        return res,nil
    }
    res.Data.Total=res.Data.Added+res.Data.Updated
    res.SetSuccess(public.LangCtx(ctx,"Central companies saved without group assignment or subscription changes"))
    return res,nil
}

func (c *ControllerV1) ListMasterOrganizations(ctx context.Context, req *v1.ListMasterOrganizationsReq) (res *v1.ListMasterOrganizationsRes, err error) {
    res=&v1.ListMasterOrganizationsRes{}
    page, pageSize:=req.Page,req.PageSize
    if page<1{page=1}
    if pageSize<1{pageSize=50}
    if pageSize>100{pageSize=100}
    search:="%"+strings.TrimSpace(req.Keyword)+"%"
    totalRows,e:=g.DB().GetAll(ctx,`SELECT COUNT(*) AS n FROM bm_master_organizations
        WHERE name ILIKE $1 OR domain ILIKE $1 OR industry ILIKE $1`,search)
    if e!=nil{res.SetError(gerror.New(e.Error()));return res,nil}
    if len(totalRows)>0{res.Data.Total=gconv.Int64(totalRows[0]["n"])}
    rows,e:=g.DB().GetAll(ctx,`SELECT id,organization_key,name,domain,website,industry,country,state_region,
        city,phone,description,metadata::text AS metadata,source_details::text AS source_details,
        created_at,updated_at FROM bm_master_organizations
        WHERE name ILIKE $1 OR domain ILIKE $1 OR industry ILIKE $1
        ORDER BY name LIMIT $2 OFFSET $3`,search,pageSize,(page-1)*pageSize)
    if e!=nil{res.SetError(gerror.New(e.Error()));return res,nil}
    res.Data.List=make([]map[string]interface{},0,len(rows))
    for _,record:=range rows{
        obj:=make(map[string]interface{},len(record))
        for k,v:=range record{obj[k]=v.Interface()}
        res.Data.List=append(res.Data.List,obj)
    }
    res.SetSuccess("Central organizations retrieved")
    return res,nil
}
