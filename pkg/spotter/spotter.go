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
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// whitelistStartPrefix in `YYYY-MM-DDT` format, can be anthing
	whitelistStartPrefix = "2000-01-01T"

	// whitelistEndPrefix in `YYYY-MM-DDT` format, has to be whitelistStartPrefix plus one day
	whitelistEndPrefix = "2000-01-02T"

	// whitelistTimePostfix in `:ssZ` format, can be anything
	whitelistTimePostfix = ":00+05:30"
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
		creationTimeStamp := node.GetCreationTimestamp()
		if _, ok := nodeAnnotations[config.SpotterExpiryTimeAnnotation]; ok {
			continue
		}
		if nodeAnnotations == nil {
			nodeAnnotations = make(map[string]string, 0)
		}
		expiryTime := ss.getExpiryTimestamp(creationTimeStamp, ss.cp.GetInt(config.SpotterTTLHours))
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
		ss.whiteListIntervals.Insert(start, end)
	}
}
func midnight(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, t.Location())
}

func add30Min(t time.Time) time.Time {
	year, month, day := t.Date()
	hour, min, sec := t.Add(30 * time.Minute).Clock()

	return time.Date(year, month, day, hour, min, sec, 0, t.Location())
}

func (ss *spotterService) getExpiryTimestamp(creationTs v1.Time, ttl int) string {

	var cet time.Time

	if ttl > 0 {
		cet = creationTs.Time.Add(time.Duration(ttl) * time.Hour)
	} else if ttl < 0 {
		panic("TTL Cannot be negative")
		// et := creationTimeUTC.Add(time.Duration(24) * time.Hour)
		// cet = et.Add(time.Duration(ttl) * time.Hour)
	}

	//We need this to compare against our whitelist intervals which is a hard-coded date + time interval
	//we dont'worry about date but time for slotting
	//once we find the ideal slot for CET we project it back to actual time

	truncatedExpiryTs := midnight(cet)
	projectedExpiryTs := whitelistStart.Add(cet.Sub(truncatedExpiryTs))

	ss.logger.Info(fmt.Sprintf("Created time %v", creationTs))
	ss.logger.Info(fmt.Sprintf("TruncatedExpiryTs %v", truncatedExpiryTs))
	ss.logger.Info(fmt.Sprintf("Calculated Expiry time %v", cet))
	ss.logger.Info(fmt.Sprintf("ProjectedExpiryTs %v", projectedExpiryTs))
	ss.logger.Info(fmt.Sprintf("Whitelist =>  %v", ss.whiteListIntervals))

	//Slot CET to bucket by sub 30 mins
	slotted := false
	//as long as the expiry time is not slotted to an available bucket loop and decrement 30 mins from exp time
	drecementedProjectedExpiry := projectedExpiryTs

	for !slotted {
		ss.whiteListIntervals.IntervalsBetween(whitelistStart, whitelistEnd, func(start, end time.Time) bool {
			if start.Before(drecementedProjectedExpiry) && end.After(drecementedProjectedExpiry) {

				slotted = true
				return false
			}
			return true
		})
		if !slotted {
			drecementedProjectedExpiry = drecementedProjectedExpiry.Add(-30 * time.Minute)
		}

	}

	slotted = false
	incrementedProjectedExpiry := projectedExpiryTs
	//as long as the expiry time is not slotted to an available bucket loop and increment 30 mins to exp time
	for !slotted {
		ss.whiteListIntervals.IntervalsBetween(whitelistStart, whitelistEnd, func(start, end time.Time) bool {
			if start.Before(incrementedProjectedExpiry) && end.After(incrementedProjectedExpiry) {

				slotted = true
				return false
			}
			return true
		})
		if !slotted {
			incrementedProjectedExpiry = add30Min(incrementedProjectedExpiry)
		}

	}

	ss.logger.Info(fmt.Sprintf("DrecementedProjectedExpiry =>  %v", drecementedProjectedExpiry))
	ss.logger.Info(fmt.Sprintf("IncrementedProjectedExpiry =>  %v", incrementedProjectedExpiry))

	var finalExpTime time.Time

	if cet.Sub(drecementedProjectedExpiry) < incrementedProjectedExpiry.Sub(cet) {
		finalExpTime = drecementedProjectedExpiry
	} else {
		finalExpTime = incrementedProjectedExpiry
	}

	//Project it back to actual date
	expTime := truncatedExpiryTs.Add(finalExpTime.Sub(whitelistStart))

	if expTime.Before(creationTs.Time) {
		expTime = expTime.Add(24 * time.Hour)
	}

	if expTime.After(creationTs.Time.Add(24 * time.Hour)) {
		expTime = expTime.Add(-24 * time.Hour)
	}

	ss.logger.Info(fmt.Sprintf("CET Slotted =>  %v", expTime.String()))
	return expTime.String()
}
