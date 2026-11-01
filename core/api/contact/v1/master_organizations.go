package v1

import (
    "billionmail-core/utility/types/api_v1"
    "github.com/gogf/gf/v2/frame/g"
)

type MasterOrganizationInput struct {
    Name          string                 `json:"name"`
    Domain        string                 `json:"domain"`
    Website       string                 `json:"website"`
    Industry      string                 `json:"industry"`
    Country       string                 `json:"country"`
    StateRegion   string                 `json:"state_region"`
    City          string                 `json:"city"`
    Phone         string                 `json:"phone"`
    Description   string                 `json:"description"`
    Metadata      map[string]interface{} `json:"metadata"`
    SourceDetails []interface{}          `json:"source_details"`
}

type ImportMasterOrganizationsReq struct {
    g.Meta        `path:"/contact/master/organizations/import" method:"post" tags:"Contact" summary:"Import independent central company records"`
    Authorization string                    `json:"authorization" in:"header"`
    Organizations []MasterOrganizationInput `json:"organizations"`
}

type ImportMasterOrganizationsRes struct {
    api_v1.StandardRes
    Data struct {
        Added int `json:"added"`
        Updated int `json:"updated"`
        Total int `json:"total"`
    } `json:"data"`
}

type ListMasterOrganizationsReq struct {
    g.Meta `path:"/contact/master/organizations/list" method:"get" tags:"Contact" summary:"List central companies"`
    Authorization string `json:"authorization" in:"header"`
    Page int `json:"page" d:"1"`
    PageSize int `json:"page_size" d:"50"`
    Keyword string `json:"keyword"`
}

type ListMasterOrganizationsRes struct {
    api_v1.StandardRes
    Data struct {
        Total int64 `json:"total"`
        List []map[string]interface{} `json:"list"`
    } `json:"data"`
}
