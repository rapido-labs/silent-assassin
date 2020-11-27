package shifter

import (
	"fmt"
	"strings"
	"time"

	"github.com/roppenlabs/silent-assassin/pkg/config"
)

type wlInterval struct {
	start time.Time
	end   time.Time
}

//initWhitelist reads the ShifterWhiteListIntervalHours from the configuration and sets whiteListIntervals
// in the shifter struct.
func (ss *ShifterService) initWhitelist() {

	ss.whiteListIntervals = []wlInterval{}
	wlStr := ss.cp.SplitStringToSlice(config.ShifterWhiteListIntervalHours, config.CommaSeparater)
	ss.logger.Info(fmt.Sprintf("Shifter: Whitelist intervals: %v", wlStr))
	for _, wl := range wlStr {
		times := strings.Split(wl, "-")
		start, err := time.Parse(timeLayout, times[0])
		if err != nil {
			ss.logger.Error(fmt.Sprintf("Shifter: Error parsing WhiteList Start time Reason: %v", err))
			panic(err)
		}

		end, err := time.Parse(timeLayout, times[1])
		if err != nil {
			ss.logger.Error(fmt.Sprintf("Shifter: Error parsing WhiteList end time Reason: %v", err))
			panic(err)
		}
		ss.whiteListIntervals = append(ss.whiteListIntervals, wlInterval{start, end})
	}
	ss.logger.Info(fmt.Sprintf("Shifter: Whitelist set initialized : %v", ss.whiteListIntervals))
}

// timeWithinWLIntervalCheck accepts 'start', 'end', 'check' times
// and returns true if 'check' is in between 'start' and 'end'
func timeWithinWLIntervalCheck(start, end, check time.Time) bool {
	if start.Before(end) {
		return !check.Before(start) && !check.After(end)
	}
	if start.Equal(end) {
		return check.Equal(start)
	}
	return check.After(start) || check.Before(end)
}
