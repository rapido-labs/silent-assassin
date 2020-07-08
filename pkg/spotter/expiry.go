package spotter

import (
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

	// whitelistEndPrefix in `YYYY-MM-DDT` format, has to be whitelistStartPrefix plus one day
	whitelistEndPrefix = "2000-01-02T"

	// whitelistTimePostfix in `:ssZ` format, can be anything
	whitelistTimePostfix = ":00Z"
)

var (
	whitelistStart time.Time
	whitelistEnd   time.Time
)

func init() {
	rand.Seed(time.Now().Unix())
	var err error

	// whitelistStart is the start of the day
	whitelistStart, err = time.Parse(time.RFC3339, whitelistStartPrefix+"00:00"+whitelistTimePostfix)
	if err != nil {
		fmt.Println(err)
		panic("whitelistStart parse error")
	}

	// whitelistEnd is the start of the next day
	whitelistEnd, err = time.Parse(time.RFC3339, whitelistEndPrefix+"00:00"+whitelistTimePostfix)
	if err != nil {
		panic("whitelistEnd parse error")
	}
}

func (ss *spotterService) initWhitelist() {
	ss.whiteListIntervals = timespanset.Empty()
	wlStr := ss.cp.GetStringSlice(config.SpotterWhiteListIntervalHours)
	for _, wl := range wlStr {
		times := strings.Split(wl, "-")
		start, err := time.Parse(time.RFC3339, whitelistStartPrefix+times[0]+whitelistTimePostfix)
		if err != nil {
			ss.logger.Error(fmt.Sprintf("Spotter: Error parsing WhiteList Start date Reason: %v", err))
			panic(err)
		}
		end, err := time.Parse(time.RFC3339, whitelistStartPrefix+times[1]+whitelistTimePostfix)
		if err != nil {
			ss.logger.Error(fmt.Sprintf("Spotter: Error parsing WhiteList End date Reason: %v", err))
			panic(err)
		}
		if end.Before(start) {
			ss.whiteListIntervals.Insert(start, whitelistEnd)
			start = whitelistStart
		}
		ss.whiteListIntervals.Insert(start, end)
	}
	ss.logger.Info(fmt.Sprintf("Spotter: Whitelist set initialized : %v", ss.whiteListIntervals))
}

// midnight returns the midnight for the date
func midnight(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, t.Location())
}

func randomNumber(a, b int) int {
	return a + rand.Intn(b-a+1)
}

func randomMinuntes(t1, t2 time.Time) time.Duration {
	interval := t2.Sub(t1).Minutes()
	randMins := randomNumber(0, int(interval))
	return time.Duration(randMins)
}

func (ss *spotterService) getExpiryTimestamp(node v1.Node) string {

	creationTsUTC := node.GetCreationTimestamp().Time.UTC()

	ss.logger.Debug(fmt.Sprintf("GetExpiryTime : Node = %v Created time = [ %v ] Created Time in UTC = [ %v ]", node.Name, node.GetCreationTimestamp(), creationTsUTC))

	truncatedCT := midnight(creationTsUTC)
	projectedCT := whitelistStart.Add(creationTsUTC.Sub(truncatedCT))

	actualExpiry := creationTsUTC.Add(24 * time.Hour)

	truncatedET := midnight(actualExpiry)
	projectedET := whitelistStart.Add(actualExpiry.Sub(truncatedET)).Add(24 * time.Hour)

	ss.logger.Debug(fmt.Sprintf("GetExpiryTime : Node = %v Projected CT = [ %v ] Projected ExpiryTime = [ %v ]", node.Name, projectedCT, projectedET))
	eligibleExpiryTimes := make([]time.Time, 0)
	for day := 0; day < 2; day++ {

		ss.whiteListIntervals.IntervalsBetween(whitelistStart, whitelistEnd, func(start, end time.Time) bool {
			if day == 1 {
				start = start.Add(24 * time.Hour)
				end = end.Add(24 * time.Hour)
			}
			ss.logger.Debug(fmt.Sprintf("GetExpiryTime : [Current Interval] Node = %v Day = %d, start = [ %v ], end = [ %v ], elegibleWLIntervals = [ %v ]", node.Name, day, start, end, eligibleExpiryTimes))
			if projectedCT.Before(start) && end.Before(projectedET) {
				timeToBeAdded := randomMinuntes(start, end)
				eligibleExpiryTimes = append(eligibleExpiryTimes, start.Add(time.Duration(timeToBeAdded)*time.Minute))

				ss.logger.Debug(fmt.Sprintf("GetExpiryTime : [Eligible Interval] Node = %v Day = %d, start = [ %v ], end = [ %v ], elegibleWLIntervals = [ %v ]", node.Name, day, start, end, eligibleExpiryTimes))
			}

			return true
		})

	}

	if len(eligibleExpiryTimes) == 0 {
		panic("Cannot find a date")
	}

	saExpirtyTime := eligibleExpiryTimes[rand.Intn(len(eligibleExpiryTimes))]

	ss.logger.Debug(fmt.Sprintf("GetExpiryTime : Node = %v elegibleWLIntervals = %v, saExpirtyTime = [ %v ]", node.Name, eligibleExpiryTimes, saExpirtyTime))
	finalexp := midnight(creationTsUTC).Add(saExpirtyTime.Sub(whitelistStart))

	if finalexp.After(actualExpiry) {
		panic("The Expiry time we calculated is after Actual Expiry Time :facepalm")
	}
	ss.logger.Debug(fmt.Sprintf("GetExpiryTime : Node = %v Final Expiry = [ %v ]", node.Name, finalexp))
	return finalexp.Format(time.RFC1123Z)
}
