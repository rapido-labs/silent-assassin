package spotter

import (
	"fmt"
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
			ss.logger.Error(fmt.Sprintf("Error parsing WhiteList Start date Reason: %v", err))
			panic(err)
		}
		end, err := time.Parse(time.RFC3339, whitelistStartPrefix+times[1]+whitelistTimePostfix)
		if err != nil {
			ss.logger.Error(fmt.Sprintf("Error parsing WhiteList End date Reason: %v", err))
			panic(err)
		}
		if end.Before(start) {
			ss.whiteListIntervals.Insert(start, whitelistEnd)
			start = whitelistStart
		}
		ss.whiteListIntervals.Insert(start, end)
	}
	ss.logger.Info(fmt.Sprintf("Whitelist set initialized : %v", ss.whiteListIntervals))
}

// midnight returns the midnight for the date
func midnight(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, t.Location())
}

// addMinToClock Increment given min to clock, but won't change date,
// To decrement set min < 0
func addMinToClock(t time.Time, min int) time.Time {
	year, month, day := t.Date()
	hour, min, sec := t.Add(time.Duration(min) * time.Minute).Clock()

	return time.Date(year, month, day, hour, min, sec, 0, t.Location())
}

//absSub Will substract two Durations and return the absolute value
func absSub(t1, t2 time.Time) time.Duration {
	if t1.After(t2) {
		return t1.Sub(t2)
	}
	return t2.Sub(t1)
}

func (ss *spotterService) getExpiryTimestamp(node v1.Node, ttl int) string {

	var cet, creationTsUTC time.Time
	creationTsUTC = node.GetCreationTimestamp().Time.UTC()
	ss.logger.Debug(fmt.Sprintf("GetExpiryTime : Node = %v Created time = [ %v ] Created Time in UTC = [ %v ]", node.Name, node.GetCreationTimestamp(), creationTsUTC))
	ss.logger.Debug(fmt.Sprintf("GetExpiryTime : TTL = %d", ttl))
	if ttl > 0 {
		cet = creationTsUTC.Add(time.Duration(ttl) * time.Hour)
	} else if ttl < 0 {
		panic("TTL Cannot be negative")
		// et := creationTimeUTC.Add(time.Duration(24) * time.Hour)
		// cet = et.Add(time.Duration(ttl) * time.Hour)
	}

	ss.logger.Debug(fmt.Sprintf("GetExpiryTime : Node = %v Calculated Expiry time (CT+TTL) before slotting = [ %v ]", node.Name, cet))

	//We need this to compare against our whitelist intervals which is a hard-coded date + time interval
	//we dont'worry about date but time for slotting
	//once we find the ideal slot for CET we project it back to actual time
	truncatedExpiryTs := midnight(cet)
	projectedExpiryTs := whitelistStart.Add(cet.Sub(truncatedExpiryTs))

	ss.logger.Debug(fmt.Sprintf("GetExpiryTime : Node = %v Truncated Expiry time = [ %v ] Projected Expiry Time [ %v ]", node.Name, truncatedExpiryTs, projectedExpiryTs))

	ch1 := make(chan time.Time)
	go ss.slotExpiryTimeToBucket(projectedExpiryTs, -30, ch1)

	ch2 := make(chan time.Time)
	go ss.slotExpiryTimeToBucket(projectedExpiryTs, 30, ch2)

	var decrementedProjectedExpiry, incrementedProjectedExpiry time.Time

	for i := 0; i < 2; i++ {
		select {
		case decrementedProjectedExpiry = <-ch1:
		case incrementedProjectedExpiry = <-ch2:
		}
	}

	ss.logger.Debug(fmt.Sprintf("GetExpiryTime : Node = %v DecrementedProjectedExpiry = [ %v ] IncrementedProjectedExpiry [ %v ]", node.Name, decrementedProjectedExpiry, incrementedProjectedExpiry))

	decrementedProjectedExpiry = truncatedExpiryTs.Add(decrementedProjectedExpiry.Sub(whitelistStart))
	incrementedProjectedExpiry = truncatedExpiryTs.Add(incrementedProjectedExpiry.Sub(whitelistStart))
	ss.logger.Debug(fmt.Sprintf("GetExpiryTime : PRORJECTED BACK Node = %v DecrementedProjectedExpiry = [ %v ] IncrementedProjectedExpiry [ %v ]", node.Name, decrementedProjectedExpiry, incrementedProjectedExpiry))
	var finalExpTime time.Time

	//Incremented Time always need not be after CET if it cycle back
	ss.logger.Debug(fmt.Sprintf("GetExpiryTime : Node = %v absSub(cet, decrementedProjectedExpiry) %v", node.Name, absSub(cet, decrementedProjectedExpiry)))
	ss.logger.Debug(fmt.Sprintf("GetExpiryTime : Node = %v absSub(incrementedProjectedExpiry, cet) %v", node.Name, absSub(incrementedProjectedExpiry, cet)))
	if absSub(cet, decrementedProjectedExpiry) < absSub(incrementedProjectedExpiry, cet) {
		finalExpTime = decrementedProjectedExpiry
	} else {
		finalExpTime = incrementedProjectedExpiry
	}

	//Project it back to actual date
	expTime := finalExpTime

	if expTime.Before(creationTsUTC) {
		ss.logger.Debug(fmt.Sprintf("GetExpiryTime : Node = %v Calculated Expiry before Creation time adding 24 Hours", node.Name))
		expTime = expTime.Add(24 * time.Hour)
	}

	if expTime.After(creationTsUTC.Add(24 * time.Hour)) {
		ss.logger.Debug(fmt.Sprintf("GetExpiryTime : Node = %v Calculated Expiry After actual Expiry sub 24 Hours", node.Name))
		expTime = expTime.Add(-24 * time.Hour)
	}

	ss.logger.Debug(fmt.Sprintf("GetExpiryTime : Node = %v Calculated Expiry time after slotting [ %v ]", node.Name, expTime))

	return expTime.Format(time.RFC1123Z)
}

func (ss *spotterService) slotExpiryTimeToBucket(projectedExpiryTs time.Time, increment int, ch chan time.Time) {
	slotted := false
	slottedProjectedExpiry := projectedExpiryTs
	//as long as the expiry time is not slotted to an available bucket loop and add increment mins (to decrement add negative) to exp time
	for !slotted {
		ss.whiteListIntervals.IntervalsBetween(whitelistStart, whitelistEnd, func(start, end time.Time) bool {
			if start.Before(slottedProjectedExpiry) && end.After(slottedProjectedExpiry) {

				slotted = true
				return false
			}
			return true
		})
		if !slotted {
			slottedProjectedExpiry = addMinToClock(slottedProjectedExpiry, increment)
		}

	}
	ch <- slottedProjectedExpiry
}
