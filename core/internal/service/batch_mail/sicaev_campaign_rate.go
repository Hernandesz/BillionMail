package batch_mail

import (
    "context"
    "fmt"
    "sync/atomic"
    "time"

    "github.com/gogf/gf/v2/database/gdb"
    "github.com/gogf/gf/v2/frame/g"
    "github.com/gogf/gf/v2/util/gconv"
)

// Conservative hard caps for all native campaign tasks combined.
// These are dispatch ATTEMPTS, not proof of mailbox acceptance or delivery.
// Avoid bursts by spacing attempts, in addition to UTC daily quotas.
const (
    globalCampaignPerHour = 20
    globalCampaignPerDay  = 200
    defaultCampaignPerHour = 20
    defaultCampaignPerDay  = 200
)

type dispatchRateState struct {
    scope    string
    lastAtMS int64
    dayKey   int
    dayCount int
    perHour  int
    perDay   int
}

func normalizeCampaignLimits(perHour, perDay int) (int, int) {
    if perHour < 1 || perHour > 60 { perHour = defaultCampaignPerHour }
    if perDay < 1 || perDay > 200 { perDay = defaultCampaignPerDay }
    return perHour, perDay
}

func utcDayKey(now time.Time) int {
    t := now.UTC()
    return t.Year()*10000 + int(t.Month())*100 + t.Day()
}

// Smallest safe wait until an attempt can be dispatched within BOTH limits.
// The interval guarantees no more than perHour attempts in any sliding hour
// (in a half-open interval) and dayCount caps all attempts in a UTC calendar day.
func dispatchDelay(now time.Time, state dispatchRateState) time.Duration {
    delay := time.Duration(0)
    today := utcDayKey(now)
    if state.dayKey == today && state.dayCount >= state.perDay {
        until := now.UTC().Truncate(24*time.Hour).Add(24*time.Hour).Sub(now.UTC())
        if until > delay { delay = until }
    }
    if state.lastAtMS > 0 {
        intervalMS := int64(time.Hour/time.Millisecond) / int64(state.perHour)
        due := time.UnixMilli(state.lastAtMS + intervalMS)
        until := due.Sub(now)
        if until > delay { delay = until }
    }
    return delay
}

func waitCampaignRate(ctx context.Context, taskID, perHour, perDay int, paused *atomic.Bool) error {
    if taskID <= 0 { return fmt.Errorf("campaign ID missing") }
    perHour, perDay = normalizeCampaignLimits(perHour,perDay)
    campaignScope := fmt.Sprintf("task:%d", taskID)

    for {
        if err := ctx.Err(); err != nil { return err }
        if paused != nil && paused.Load() {
            if err := rateSleep(ctx,time.Second); err != nil { return err }
            continue
        }

        delay, err := reserveCampaignAttempt(ctx,campaignScope,perHour,perDay)
        if err != nil {
            // Fail closed: do not send on PostgreSQL errors or missing limiter state.
            return fmt.Errorf("campaign rate limiter unavailable: %w",err)
        }
        if delay <= 0 {
            // A permit is consumed BEFORE attempting SMTP delivery. Failures count
            // against quotas deliberately, avoiding a rapid retry storm.
            return nil
        }
        // Short polling permits pauses and cancellations without long sleeps.
        if delay > 5*time.Second { delay = 5*time.Second }
        if err := rateSleep(ctx,delay); err != nil { return err }
    }
}

func rateSleep(ctx context.Context, delay time.Duration) error {
    timer := time.NewTimer(delay)
    defer timer.Stop()
    select {
    case <-ctx.Done(): return ctx.Err()
    case <-timer.C: return nil
    }
}

// All campaign senders acquire the SAME PostgreSQL transaction lock, ensuring
// no concurrency races across campaigns, processes, replicas or service restarts.
func reserveCampaignAttempt(ctx context.Context, campaignScope string, perHour, perDay int) (time.Duration,error) {
    var retry time.Duration
    err := g.DB().Transaction(ctx,func(ctx context.Context, tx gdb.TX) error {
        if _, err := tx.GetOne("SELECT pg_advisory_xact_lock(6619, 1001)"); err != nil { return err }
        if _, err := tx.Exec(`INSERT INTO bm_sicaev_campaign_rate(scope)
            VALUES ('global'), (?) ON CONFLICT (scope) DO NOTHING`,campaignScope); err != nil { return err }

        now := time.Now().UTC()
        today := utcDayKey(now)
        rows := []dispatchRateState{
            {scope:"global",perHour:globalCampaignPerHour,perDay:globalCampaignPerDay},
            {scope:campaignScope,perHour:perHour,perDay:perDay},
        }
        for i := range rows {
            record, err := tx.GetOne(`SELECT last_at_ms,day_key,day_count
                FROM bm_sicaev_campaign_rate WHERE scope=? FOR UPDATE`,rows[i].scope)
            if err != nil { return err }
            if record == nil { return fmt.Errorf("rate state missing: %s",rows[i].scope) }
            rows[i].lastAtMS = gconv.Int64(record["last_at_ms"])
            rows[i].dayKey = gconv.Int(record["day_key"])
            rows[i].dayCount = gconv.Int(record["day_count"])
            if wait := dispatchDelay(now,rows[i]); wait > retry { retry = wait }
        }
        if retry > 0 { return nil }
        for _, row := range rows {
            count := 1
            if row.dayKey == today { count = row.dayCount + 1 }
            if _, err := tx.Exec(`UPDATE bm_sicaev_campaign_rate
                SET last_at_ms=?,day_key=?,day_count=? WHERE scope=?`,
                now.UnixMilli(),today,count,row.scope); err != nil { return err }
        }
        return nil
    })
    return retry,err
}
