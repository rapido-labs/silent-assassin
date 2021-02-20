package spotter

import (
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/google/go-intervals/timespanset"
	"github.com/roppenlabs/silent-assassin/pkg/config"
	v1 "k8s.io/api/core/v1"
)

const (
	// whitelistStartPrefix in `YYYY-MM-DDT` format, can be anthing
	whitelistStartPrefix = "2000-01-01T"

	// whitelistTimePostfix in `:ssZ` format, can be anything
	whitelistTimePostfix = ":00Z"
)

var (
	whitelistStart = constructTruncatedTime("00:00")
)

func init() {
	rand.Seed(time.Now().Unix())
}

func (ss *spotterService) initWhitelist(whitelistIntervals string) {
	ss.whiteListIntervals = timespanset.Empty()
	wlStr := strings.Split(whitelistIntervals, config.CommaSeparater)
	for _, wl := range wlStr {
		times := strings.Split(wl, "-")
		start := constructTruncatedTime(times[0])
		end := constructTruncatedTime(times[1])

		if start.Before(end) {
			// if start time is before end time, add this normal time range
			ss.whiteListIntervals.Insert(start, end)

			// also add the same time range, but in the following day, this is helpful for
			// finding interval intersections between node lifecycle and whitelist intervals
			ss.whiteListIntervals.Insert(start.Add(time.Hour*24), end.Add(time.Hour*24))
		} else {
			// if end time is after **or equal** to start time, it indicates end time
			// is in the following day of start time, add a 24 hour to end time
			ss.whiteListIntervals.Insert(start, end.Add(time.Hour*24))

			// same as before, add the same interval in the following day
			ss.whiteListIntervals.Insert(start.Add(time.Hour*24), end.Add(time.Hour*48))
		}
	}
	ss.logger.Info(fmt.Sprintf("Spotter: Whitelist set initialized : %v", ss.whiteListIntervals))
}

func constructTruncatedTime(timeString string) time.Time {
	fullTimeString := whitelistStartPrefix + timeString + whitelistTimePostfix
	t, err := time.Parse(time.RFC3339, fullTimeString)
	if err != nil {
		panic(fmt.Sprintf("parseTimeInDay parse error: %v", err))
	}
	return t
}

func truncateTime(t time.Time) time.Time {
	return time.Date(
		whitelistStart.Year(),
		whitelistStart.Month(),
		whitelistStart.Day(),
		t.Hour(),
		t.Minute(),
		t.Second(),
		t.Nanosecond(),
		t.Location(),
	)
}

// randomTimeBetween returns a timestamp between t1 and t2: [t1, t2)
func randomTimeBetween(t1, t2 time.Time) time.Time {
	minutes := int(t2.Sub(t1).Minutes())
	randMinutes := rand.Intn(minutes)
	return t1.Add(time.Duration(randMinutes * int(time.Minute)))
}

func minTime(t1, t2 time.Time) time.Time {
	if t1.Before(t2) {
		return t1
	}
	return t2
}

func maxTime(t1, t2 time.Time) time.Time {
	if t1.Before(t2) {
		return t2
	}
	return t1
}

func (ss *spotterService) getExpiryTimestamp(node v1.Node) (string, error) {

	creationTsUTC := node.GetCreationTimestamp().Time.UTC()

	ss.logger.Debug(fmt.Sprintf("GetExpiryTime : Node = %v Created time = [ %v ] Created Time in UTC = [ %v ]", node.Name, node.GetCreationTimestamp(), creationTsUTC))

	// truncate year/month/day part of creation and expiration time to be on the same day of whitelist intervals
	projectedCT := truncateTime(creationTsUTC)
	projectedET := projectedCT.Add(24 * time.Hour)
	ss.logger.Debug(fmt.Sprintf("GetExpiryTime : Node = %v Projected CT = [ %v ] Projected ExpiryTime = [ %v ]", node.Name, projectedCT, projectedET))

	// enumerate every time interval that falls between projected creation and projected node expiration
	// all of these intervals are eligible candidates for killing this node
	// randomly pick one of these interval and pick a random time in that interval as selected expieration of this node
	eligibleExpiryTimes := make([]time.Time, 0)
	ss.whiteListIntervals.IntervalsBetween(projectedCT, projectedET, func(start, end time.Time) bool {
		expiryTime := randomTimeBetween(start, end)
		eligibleExpiryTimes = append(eligibleExpiryTimes, expiryTime)
		ss.logger.Debug(fmt.Sprintf("GetExpiryTime : [Eligible Interval] Node = %v start = [ %v ], end = [ %v ], elegibleWLIntervals = [ %v ]", node.Name, start, end, eligibleExpiryTimes))

		return true
	})

	if len(eligibleExpiryTimes) == 0 {
		return "", errors.New("Cannot find a date")
	}

	saExpirtyTime := eligibleExpiryTimes[rand.Intn(len(eligibleExpiryTimes))]
	ss.logger.Debug(fmt.Sprintf("GetExpiryTime : Node = %v elegibleWLIntervals = %v, saExpirtyTime = [ %v ]", node.Name, eligibleExpiryTimes, saExpirtyTime))

	// project this expiration time back to the day of node creation
	finalexp := creationTsUTC.Add(saExpirtyTime.Sub(projectedCT))

	actualExpiry := creationTsUTC.Add(24 * time.Hour)
	if finalexp.After(actualExpiry) {
		return "", errors.New("The Expiry time we calculated is after Actual Expiry Time :facepalm")
	}
	finalexp = finalexp.In(node.GetCreationTimestamp().Time.Location())
	ss.logger.Debug(fmt.Sprintf("GetExpiryTime : Node = %v Final Expiry = [ %v ]", node.Name, finalexp))
	return finalexp.Format(time.RFC1123Z), nil
}
