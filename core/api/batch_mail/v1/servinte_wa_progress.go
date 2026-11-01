package v1

import (
    "billionmail-core/utility/types/api_v1"
    "github.com/gogf/gf/v2/frame/g"
)

// Monitoring endpoint is read-only: it cannot send, pause or modify campaigns.
// A recipient is "processed" even if its SMTP attempt failed; delivery is
// separately measured by the asynchronous mail statistics.
type ServinteWAProgressReq struct {
    g.Meta `path:"/batch_mail/task/servinte_wa_progress" method:"get" tags:"BatchMail" summary:"Read SERVINTE campaign progress for WhatsApp notifications"`
    Authorization string `json:"authorization" in:"header"`
}

type ServinteWATaskProgress struct {
    Id           int    `json:"id"`
    Subject      string `json:"subject"`
    GroupId      int    `json:"group_id"`
    TemplateId   int    `json:"template_id"`
    StartTime    int64  `json:"start_time"`
    TaskProcess  int    `json:"task_process"`
    Pause        int    `json:"pause"`
    Total        int    `json:"total"`
    Processed    int    `json:"processed"`
    Pending      int    `json:"pending"`
    InFlight     int    `json:"in_flight"`
    SentLogged   int    `json:"sent_logged"`
    Delivered    int    `json:"delivered"`
    Bounced      int    `json:"bounced"`
    Deferred     int    `json:"deferred"`
    MaxPerHour   int    `json:"max_per_hour"`
    MaxPerDay    int    `json:"max_per_day"`
}

type ServinteWAProgressRes struct {
    api_v1.StandardRes
    Data struct {
        Tasks []ServinteWATaskProgress `json:"tasks"`
    } `json:"data"`
}
