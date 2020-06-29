package spotter

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/go-intervals/timespanset"
	"github.com/roppenlabs/silent-assassin/pkg/config"
	"github.com/roppenlabs/silent-assassin/pkg/k8s"
	"github.com/roppenlabs/silent-assassin/pkg/logger"
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

type spotterService struct {
	cp                 config.IProvider
	logger             logger.IZapLogger
	kubeClient         k8s.IKubernetesClient
	whiteListIntervals *timespanset.Set
}

func NewSpotterService(cp config.IProvider, zl logger.IZapLogger, kc k8s.IKubernetesClient) spotterService {
	return spotterService{
		cp:         cp,
		logger:     zl,
		kubeClient: kc,
	}
}

func (ss spotterService) Start(ctx context.Context, wg *sync.WaitGroup) {

	ss.logger.Info(fmt.Sprintf("Starting Spotter Loop - Poll Interval : %d", ss.cp.GetInt(config.SpotterPollIntervalMs)))

	ss.initWhitelist()

	for {
		select {
		case <-ctx.Done():
			ss.logger.Info("Shutting down spotter service")
			wg.Done()
			return
		default:
			ss.spot()
			time.Sleep(time.Millisecond * time.Duration(ss.cp.GetInt(config.SpotterPollIntervalMs)))
		}
	}
}

func (ss spotterService) spot() {

	nodes := ss.kubeClient.GetNodes(ss.cp.GetStringSlice(config.SpotterNodeSelectors))

	ss.logger.Debug(fmt.Sprintf("Fetched %d node(s)", len(nodes.Items)))

	for _, node := range nodes.Items {
		nodeAnnotations := node.GetAnnotations()

		if _, ok := nodeAnnotations[config.SpotterExpiryTimeAnnotation]; ok {
			continue
		}
		if nodeAnnotations == nil {
			nodeAnnotations = make(map[string]string, 0)
		}
		expiryTime := ss.getExpiryTimestamp(node, ss.cp.GetInt(config.SpotterTTLHours))
		ss.logger.Debug(fmt.Sprintf("spot() : Node = %v Creation Time = [ %v ] Expirty Time [ %v ]", node.Name, node.GetCreationTimestamp(), expiryTime))
		nodeAnnotations[config.SpotterExpiryTimeAnnotation] = expiryTime

		node.SetAnnotations(nodeAnnotations)
		err := ss.kubeClient.AnnotateNode(node)
		if err != nil {
			ss.logger.Error(fmt.Sprintf("Failed to annotate node : %s", node.ObjectMeta.Name))
			panic(err)
		}
		ss.logger.Info(fmt.Sprintf("Annotated node : %s", node.ObjectMeta.Name))
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

func (ss *spotterService) getExpiryTimestamp(node v1.Node, ttl int) string {

	var cet, creationTsUTC time.Time
	creationTsUTC = node.GetCreationTimestamp().Time.UTC()
	ss.logger.Debug(fmt.Sprintf("GetExpiryTime : Node = %v Created time = [ %v ] Created Time in UTC = [ %v ]", node.Name, node.GetCreationTimestamp(), creationTsUTC))
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

	var finalExpTime time.Time

	if cet.Sub(decrementedProjectedExpiry) < incrementedProjectedExpiry.Sub(cet) {
		finalExpTime = decrementedProjectedExpiry
	} else {
		finalExpTime = incrementedProjectedExpiry
	}

	//Project it back to actual date
	expTime := truncatedExpiryTs.Add(finalExpTime.Sub(whitelistStart))
	ss.logger.Debug(fmt.Sprintf("GetExpiryTime : Node = %v Calculated Expiry time afte slotting [ %v ]", node.Name, expTime))

	if expTime.Before(creationTsUTC) {
		ss.logger.Debug(fmt.Sprintf("GetExpiryTime : Node = %v Calculated Expiry before Creation time adding 24 Hours", node.Name))
		expTime = expTime.Add(24 * time.Hour)
	}

	if expTime.After(creationTsUTC.Add(24 * time.Hour)) {
		ss.logger.Debug(fmt.Sprintf("GetExpiryTime : Node = %v Calulated Expiry After actual Expiry sub 24 Hours", node.Name))
		expTime = expTime.Add(-24 * time.Hour)
	}

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
