package batch_mail

import (
    "billionmail-core/api/batch_mail/v1"
    "billionmail-core/internal/service/public"
    "context"

    "github.com/gogf/gf/v2/errors/gerror"
    "github.com/gogf/gf/v2/frame/g"
    "github.com/gogf/gf/v2/util/gconv"
)

// ServinteWAProgress is read-only. It reports attempts processed in recipient_info,
// which include both successful and failed attempts. "Delivered" is separately
// sourced from delayed tracking statistics and may lag the sending progress.
// The task identity guards avoid reporting a newly reused ID by mistake.
func (c *ControllerV1) ServinteWAProgress(ctx context.Context, req *v1.ServinteWAProgressReq) (res *v1.ServinteWAProgressRes, err error) {
    res = &v1.ServinteWAProgressRes{}
    rows, e := g.DB().GetAll(ctx, `
        SELECT t.id, t.subject, t.group_id, t.template_id, t.start_time,
               t.task_process, t.pause, t.max_per_hour, t.max_per_day,
               COUNT(ri.id)::integer AS total,
               COUNT(ri.id) FILTER (WHERE ri.is_sent = 1)::integer AS processed,
               COUNT(ri.id) FILTER (WHERE ri.is_sent = 0)::integer AS pending,
               COUNT(ri.id) FILTER (WHERE ri.is_sent = 2)::integer AS in_flight,
               COALESCE(t.sends_count, 0) AS sent_logged,
               COALESCE(t.delivered_count, 0) AS delivered,
               COALESCE(t.bounced_count, 0) AS bounced,
               COALESCE(t.deferred_count, 0) AS deferred
        FROM email_tasks t LEFT JOIN recipient_info ri ON ri.task_id=t.id
        WHERE (t.id=1 AND t.group_id=7 AND t.template_id=3 AND t.start_time=1793714400)
           OR (t.id=2 AND t.group_id=8 AND t.template_id=4 AND t.start_time=1793714400)
        GROUP BY t.id
        ORDER BY t.id
    `)
    if e != nil {
        res.SetError(gerror.New("Unable to read native campaign progress: "+e.Error()))
        return res,nil
    }
    res.Data.Tasks=make([]v1.ServinteWATaskProgress,0,len(rows))
    for _, r := range rows {
        p := v1.ServinteWATaskProgress{
            Id: gconv.Int(r["id"]),
            Subject: gconv.String(r["subject"]),
            GroupId: gconv.Int(r["group_id"]),
            TemplateId: gconv.Int(r["template_id"]),
            StartTime: gconv.Int64(r["start_time"]),
            TaskProcess: gconv.Int(r["task_process"]),
            Pause: gconv.Int(r["pause"]),
            Total: gconv.Int(r["total"]),
            Processed: gconv.Int(r["processed"]),
            Pending: gconv.Int(r["pending"]),
            InFlight: gconv.Int(r["in_flight"]),
            SentLogged: gconv.Int(r["sent_logged"]),
            Delivered: gconv.Int(r["delivered"]),
            Bounced: gconv.Int(r["bounced"]),
            Deferred: gconv.Int(r["deferred"]),
            MaxPerHour: gconv.Int(r["max_per_hour"]),
            MaxPerDay: gconv.Int(r["max_per_day"]),
        }
        res.Data.Tasks=append(res.Data.Tasks,p)
    }
    res.SetSuccess(public.LangCtx(ctx,"Campaign progress ready"))
    return res,nil
}
