package batch_mail

import (
    "testing"
    "time"
)

// No SMTP delivery is performed by these tests.
func TestCampaignRatePresetsAndGlobalFloor(t *testing.T) {
    now := time.Date(2026, 11, 1, 12, 0, 0, 0, time.UTC)
    last := now.Add(-time.Minute).UnixMilli()
    scenarios := []struct {
        name string
        perHour int
        expected time.Duration
    }{
        {"20 per hour",20,2*time.Minute},
        {"5 per hour",5,11*time.Minute},
        {"1 per hour",1,59*time.Minute},
        {"1 per minute",60,0},
    }
    for _, tc := range scenarios {
        t.Run(tc.name,func(t *testing.T) {
            s:=dispatchRateState{lastAtMS:last,dayKey:utcDayKey(now),perHour:tc.perHour,perDay:200}
            if got:=dispatchDelay(now,s); got!=tc.expected {
                t.Fatalf("delay=%v want=%v",got,tc.expected)
            }
        })
    }
    global:=dispatchRateState{lastAtMS:last,dayKey:utcDayKey(now),perHour:globalCampaignPerHour,perDay:globalCampaignPerDay}
    perTask:=dispatchRateState{lastAtMS:last,dayKey:utcDayKey(now),perHour:60,perDay:200}
    if dispatchDelay(now,global) < 2*time.Minute || dispatchDelay(now,perTask) != 0 {
        t.Fatal("1/minute campaign may not override 20/hour global limiter")
    }
}

func TestCampaignDailyLimitAndReset(t *testing.T) {
    today:=time.Date(2026,11,1,23,59,0,0,time.UTC)
    full:=dispatchRateState{dayKey:utcDayKey(today),dayCount:200,perHour:20,perDay:200}
    if wait:=dispatchDelay(today,full);wait!=time.Minute {
        t.Fatalf("daily limiter should wait for next UTC midnight, got %v",wait)
    }
    tomorrow:=today.Add(2*time.Minute)
    if wait:=dispatchDelay(tomorrow,full);wait!=0 {
        t.Fatalf("daily quota should reset on next UTC day, got %v",wait)
    }
    limited:=dispatchRateState{dayKey:utcDayKey(today),dayCount:20,perHour:20,perDay:20}
    if wait:=dispatchDelay(today,limited);wait!=time.Minute {
        t.Fatalf("20/day campaign should honor local cap: %v",wait)
    }
}

func TestNormalizeCampaignLimits(t *testing.T) {
    cases:=[]struct{h,d,eh,ed int}{
        {0,0,20,200},{1,20,1,20},{60,100,60,100},{1000,-5,20,200},
    }
    for _,c:=range cases {
        h,d:=normalizeCampaignLimits(c.h,c.d)
        if h!=c.eh||d!=c.ed { t.Fatalf("got %d %d expected %d %d",h,d,c.eh,c.ed) }
    }
}
