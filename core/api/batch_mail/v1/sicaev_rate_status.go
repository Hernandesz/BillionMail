package v1

import (
    "billionmail-core/utility/types/api_v1"
    "github.com/gogf/gf/v2/frame/g"
)

type CampaignRateStatusReq struct {
    g.Meta `path:"/batch_mail/task/rate_status" method:"get" tags:"BatchMail" summary:"Read global native campaign rate policy"`
    Authorization string `json:"authorization" in:"header"`
}
type CampaignRateStatusRes struct {
    api_v1.StandardRes
    Data struct {
        GlobalPerHour int `json:"global_per_hour"`
        GlobalPerDay int `json:"global_per_day"`
        TodayAttempts int `json:"today_attempts"`
        TodayKey int `json:"today_key_utc"`
        NextAllowedUnixMS int64 `json:"next_allowed_unix_ms"`
        SchemaReady bool `json:"schema_ready"`
    } `json:"data"`
}
