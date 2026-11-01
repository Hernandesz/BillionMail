package v1

import (
    "billionmail-core/utility/types/api_v1"
    "github.com/gogf/gf/v2/frame/g"
)

// Adds CRM directory entries only; never creates a subscription or group membership.
type MasterContactInput struct {
    Email         string                 `json:"email"`
    FullName      string                 `json:"full_name"`
    JobTitle      string                 `json:"job_title"`
    Department    string                 `json:"department"`
    CompanyName   string                 `json:"company_name"`
    Industry      string                 `json:"industry"`
    Phone         string                 `json:"phone"`
    Website       string                 `json:"website"`
    LinkedInURL   string                 `json:"linkedin_url"`
    City          string                 `json:"city"`
    StateRegion   string                 `json:"state_region"`
    Country       string                 `json:"country"`
    Metadata      map[string]interface{} `json:"metadata"`
    SourceDetails []interface{}          `json:"source_details"`
}

type ImportMasterContactsReq struct {
    g.Meta        `path:"/contact/master/import" method:"post" tags:"Contact" summary:"Save contacts to central directory without subscribing"`
    Authorization string               `json:"authorization" in:"header"`
    Contacts      []MasterContactInput `json:"contacts" v:"required"`
}

type ImportMasterContactsRes struct {
    api_v1.StandardRes
    Data struct {
        Added   int `json:"added"`
        Updated int `json:"updated"`
        Total   int `json:"total"`
    } `json:"data"`
}
