package database_initialization

import (
    "context"

    "github.com/gogf/gf/v2/frame/g"
)

// Repair only the known SERVINTE future campaign created before the tx-count
// fix. An exact match prevents touching unrelated campaigns or past sends.
// No campaign is started, paused or otherwise modified by this reconciliation.
func init() {
    registerHandler(func() {
        ctx := context.Background()
        const sql = `UPDATE email_tasks AS t
            SET recipient_count = 89
            WHERE t.id = 1
              AND t.group_id = 7
              AND t.template_id = 3
              AND t.start_time = 1793714400
              AND t.task_process = 0
              AND t.recipient_count = 0
              AND (SELECT COUNT(*) FROM recipient_info AS ri
                   WHERE ri.task_id = t.id AND ri.is_sent = 0) = 89`
        r,err := g.DB().Exec(ctx,sql)
        if err != nil {
            g.Log().Error(ctx,"SICAEV first future campaign count repair: ",err)
            return
        }
        if changed,err:=r.RowsAffected();err==nil {
            g.Log().Infof(ctx,"SICAEV first future campaign recipient count reconciled: %d row(s)",changed)
        }
    })
}
