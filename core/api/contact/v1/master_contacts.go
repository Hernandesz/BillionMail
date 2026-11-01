package v1

import (
    "billionmail-core/utility/types/api_v1"
    "github.com/gogf/gf/v2/frame/g"
)

// Master contacts are a CRM directory, not a subscription list.
// Subscription and verification are never inferred from a public email address.
type ListMasterContactsReq struct {
    g.Meta        `path:"/contact/master/list" method:"get" tags:"Contact" summary:"Central directory of all contact emails"`
    Authorization string `json:"authorization" in:"header"`
    Page          int    `json:"page" d:"1"`
    PageSize      int    `json:"page_size" d:"20"`
    Keyword       string `json:"keyword"`
    Filter        string `json:"filter" d:"all"`
}

type MasterGroupInfo struct {
    Id   int    `json:"id"`
    Name string `json:"name"`
}

type MasterTagInfo struct {
    Id   int    `json:"id"`
    Name string `json:"name"`
}

type MasterContactInfo struct {
    Id                 int64                  `json:"id"`
    Email              string                 `json:"email"`
    FullName           string                 `json:"full_name"`
    JobTitle           string                 `json:"job_title"`
    Department         string                 `json:"department"`
    CompanyName        string                 `json:"company_name"`
    Industry           string                 `json:"industry"`
    Phone              string                 `json:"phone"`
    Website            string                 `json:"website"`
    LinkedInURL        string                 `json:"linkedin_url"`
    City               string                 `json:"city"`
    StateRegion        string                 `json:"state_region"`
    Country            string                 `json:"country"`
    VerificationStatus string                 `json:"verification_status"`
    ConsentStatus      string                 `json:"consent_status"`
    NativeStatus       string                 `json:"native_status"`
    Source             string                 `json:"source"`
    UpdatedAt          int64                  `json:"updated_at"`
    Groups             []MasterGroupInfo      `json:"groups"`
    Tags               []MasterTagInfo        `json:"tags"`
    Metadata           map[string]interface{} `json:"metadata"`
    SourceDetails      []interface{}          `json:"source_details"`
}

type ListMasterContactsRes struct {
    api_v1.StandardRes
    Data struct {
        Total int64               `json:"total"`
        List  []*MasterContactInfo `json:"list"`
    } `json:"data"`
}
