package batch_mail

import (
    "billionmail-core/api/batch_mail/v1"
    "billionmail-core/internal/service/public"
    "context"
    "time"

    "github.com/gogf/gf/v2/errors/gerror"
    "github.com/gogf/gf/v2/frame/g"
    "github.com/gogf/gf/v2/util/gconv"
)

// CampaignRateStatus is intentionally read-only and never reserves a send slot.
func (c *ControllerV1) CampaignRateStatus(ctx context.Context, req *v1.CampaignRateStatusReq) (res *v1.CampaignRateStatusRes, err error) {
    res = &v1.CampaignRateStatusRes{}
    row, queryErr := g.DB().GetOne(ctx,
        "SELECT last_at_ms,day_key,day_count FROM bm_sicaev_campaign_rate WHERE scope='global'")
    if queryErr != nil || row == nil {
        res.SetError(gerror.New("Campaign rate state is unavailable"))
        return res, nil
    }
    now := time.Now().UTC()
    dayKey := now.Year()*10000 + int(now.Month())*100 + now.Day()
    attempts := 0
    if gconv.Int(row["day_key"]) == dayKey {
        attempts = gconv.Int(row["day_count"])
    }
    res.Data.GlobalPerHour = 20
    res.Data.GlobalPerDay = 200
    res.Data.TodayAttempts = attempts
    res.Data.TodayKey = dayKey
    res.Data.NextAllowedUnixMS = gconv.Int64(row["last_at_ms"]) + 180000
    res.Data.SchemaReady = true
    res.SetSuccess(public.LangCtx(ctx,"Campaign rate controls available"))
    return res,nil
}
