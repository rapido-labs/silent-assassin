package spotter

import (
	"fmt"
	"time"

	"github.com/roppenlabs/silent-assassin/pkg/config"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ExpiryTestData struct {
	Name               string
	CreationTime       time.Time
	EligibleWLs        []TimeSpan
	WhitelistIntervals string
}
type TimeSpan struct {
	Start time.Time
	End   time.Time
}

func parseTime(t string) time.Time {
	r, err := time.Parse(time.RFC1123Z, t)
	if err != nil {
		panic(err)
	}
	return r
}

func verifyNodeExpiry(t time.Time, eligibleWLs []TimeSpan) bool {
	for _, wl := range eligibleWLs {
		if (t.After(wl.Start) || t.Equal(wl.Start)) && (t.Before(wl.End) || t.Equal(wl.End)) {
			return true
		}
	}
	return false
}

//If the CET (CT+TTL) falls in a WhiteList Interval , it should be used as is
func (suite *SpotterTestSuite) TestShouldSlotNodeExpTimeToOneOfElegibleWLInRandom() {
	testData := []ExpiryTestData{
		{
			Name:               "Node-1",
			CreationTime:       parseTime("Mon, 22 Jun 2020 10:10:00 +0000"),
			WhitelistIntervals: "00:00-06:00,12:00-14:00",
			EligibleWLs: []TimeSpan{
				{
					Start: parseTime("Mon, 22 Jun 2020 12:00:00 +0000"),
					End:   parseTime("Mon, 22 Jun 2020 14:00:00 +0000"),
				},
				{
					Start: parseTime("Mon, 23 Jun 2020 00:00:00 +0000"),
					End:   parseTime("Mon, 23 Jun 2020 06:00:00 +0000"),
				},
			},
		},
		{
			Name:               "Node-2",
			CreationTime:       parseTime("Mon, 22 Jun 2020 15:40:00 +0000"),
			WhitelistIntervals: "00:00-06:00,12:00-14:00",
			EligibleWLs: []TimeSpan{
				{
					Start: parseTime("Mon, 23 Jun 2020 00:00:00 +0000"),
					End:   parseTime("Mon, 23 Jun 2020 06:00:00 +0000"),
				},
				{
					Start: parseTime("Mon, 23 Jun 2020 12:00:00 +0000"),
					End:   parseTime("Mon, 23 Jun 2020 14:00:00 +0000"),
				},
			},
		},
		{
			Name:               "Node-3",
			CreationTime:       parseTime("Mon, 22 Jun 2020 00:40:00 +0000"),
			WhitelistIntervals: "00:00-06:00,12:00-14:00",
			EligibleWLs: []TimeSpan{
				{
					Start: parseTime("Mon, 22 Jun 2020 00:40:00 +0000"),
					End:   parseTime("Mon, 22 Jun 2020 06:00:00 +0000"),
				},
				{
					Start: parseTime("Mon, 22 Jun 2020 12:00:00 +0000"),
					End:   parseTime("Mon, 22 Jun 2020 14:00:00 +0000"),
				},
				{
					Start: parseTime("Mon, 23 Jun 2020 00:00:00 +0000"),
					End:   parseTime("Mon, 23 Jun 2020 00:40:00 +0000"),
				},
			},
		},
		{
			Name:               "Node-4",
			CreationTime:       parseTime("Mon, 22 Jun 2020 22:20:00 +0000"),
			WhitelistIntervals: "00:00-06:00,12:00-14:00",
			EligibleWLs: []TimeSpan{
				{
					Start: parseTime("Mon, 23 Jun 2020 00:00:00 +0000"),
					End:   parseTime("Mon, 23 Jun 2020 06:00:00 +0000"),
				},
				{
					Start: parseTime("Mon, 23 Jun 2020 12:00:00 +0000"),
					End:   parseTime("Mon, 23 Jun 2020 14:00:00 +0000"),
				},
			},
		},
		{
			Name:               "Node-5",
			CreationTime:       parseTime("Thu, 18 Feb 2021 18:58:52 +0000"),
			WhitelistIntervals: "17:00-00:00",
			EligibleWLs: []TimeSpan{
				{
					Start: parseTime("Thu, 18 Feb 2021 18:58:52 +0000"),
					End:   parseTime("Thu, 19 Feb 2021 00:00:00 +0000"),
				},
				{
					Start: parseTime("Thu, 19 Feb 2021 17:00:00 +0000"),
					End:   parseTime("Thu, 19 Feb 2021 18:58:52 +0000"),
				},
			},
		},
		{
			Name:               "whitelist full day 1",
			CreationTime:       parseTime("Thu, 18 Feb 2021 18:58:52 +0000"),
			WhitelistIntervals: "00:00-00:00",
			EligibleWLs: []TimeSpan{
				{
					Start: parseTime("Thu, 18 Feb 2021 18:58:52 +0000"),
					End:   parseTime("Thu, 19 Feb 2021 18:58:52 +0000"),
				},
			},
		},
		{
			Name:               "whitelist full day 2",
			CreationTime:       parseTime("Thu, 18 Feb 2021 01:00:00 +0000"),
			WhitelistIntervals: "01:00-01:00",
			EligibleWLs: []TimeSpan{
				{
					Start: parseTime("Thu, 18 Feb 2021 01:00:00 +0000"),
					End:   parseTime("Thu, 19 Feb 2021 01:00:00 +0000"),
				},
			},
		},
		{
			Name:               "very short whitelist",
			CreationTime:       parseTime("Thu, 18 Feb 2021 01:00:00 +0000"),
			WhitelistIntervals: "02:00-02:01,03:00-03:01",
			EligibleWLs: []TimeSpan{
				{
					Start: parseTime("Thu, 18 Feb 2021 02:00:00 +0000"),
					End:   parseTime("Thu, 18 Feb 2021 02:01:00 +0000"),
				},
				{
					Start: parseTime("Thu, 18 Feb 2021 03:00:00 +0000"),
					End:   parseTime("Thu, 18 Feb 2021 03:01:00 +0000"),
				},
			},
		},
		{
			Name:               "duplicated whitelist 1",
			CreationTime:       parseTime("Thu, 18 Feb 2021 01:00:00 +0000"),
			WhitelistIntervals: "02:00-03:00,02:00-03:00,02:00-03:00",
			EligibleWLs: []TimeSpan{
				{
					Start: parseTime("Thu, 18 Feb 2021 02:00:00 +0000"),
					End:   parseTime("Thu, 18 Feb 2021 03:00:00 +0000"),
				},
			},
		},
		{
			Name:               "duplicated whitelist 2",
			CreationTime:       parseTime("Thu, 18 Feb 2021 01:00:00 +0000"),
			WhitelistIntervals: "02:00-03:00,01:00-04:00",
			EligibleWLs: []TimeSpan{
				{
					Start: parseTime("Thu, 18 Feb 2021 01:00:00 +0000"),
					End:   parseTime("Thu, 18 Feb 2021 04:00:00 +0000"),
				},
			},
		},
	}
	for _, testInput := range testData {
		testInput := testInput
		// since the chosen expiration time is randomized, we run each test a few more times to ensure
		for idx := 0; idx < 10; idx++ {
			suite.Run(fmt.Sprintf("%s-%d", testInput.Name, idx), func() {
				config := config.InitValue(suite.defaultConfigValues(), map[string]interface{}{
					config.SpotterWhiteListIntervalHours: testInput.WhitelistIntervals,
				})

				ss := NewSpotterService(config, suite.logger, suite.k8sMock, suite.notifierMock)
				nodeToBeAnnotated := v1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name:              testInput.Name,
						CreationTimestamp: metav1.NewTime(testInput.CreationTime),
						Annotations:       map[string]string{"node.alpha.kubernetes.io/ttl": "0"}}}

				saExpTimeString, err := ss.getExpiryTimestamp(nodeToBeAnnotated)
				suite.Assert().NoError(err)

				saExpTime, err := time.Parse(time.RFC1123Z, saExpTimeString)
				suite.Assert().NoError(err)

				suite.Assert().True(verifyNodeExpiry(saExpTime, testInput.EligibleWLs),
					"SA_Expiry time =[ %v ] didn't fall within one of the eligible WL interval = [ %v ] for Node = %v",
					saExpTime, testInput.EligibleWLs, testInput.Name)
			})
		}
	}
}

func (suite *SpotterTestSuite) TestShouldReturnETinSameTimeZoneAsCT() {

	ss := NewSpotterService(suite.config, suite.logger, suite.k8sMock, suite.notifierMock)
	creationTime := parseTime("Mon, 22 Jun 2020 22:20:00 +0530")
	nodeToBeAnnotated := v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "Node-IST",
			CreationTimestamp: metav1.NewTime(creationTime),
			Annotations:       map[string]string{"node.alpha.kubernetes.io/ttl": "0"}}}
	saExpTimeString, _ := ss.getExpiryTimestamp(nodeToBeAnnotated)
	saExpTime, _ := time.Parse(time.RFC1123Z, saExpTimeString)

	suite.Assert().Equal(saExpTime.Location(), creationTime.Location(), "CT and ET TimeZone does not match")
}

func (suite *SpotterTestSuite) TestRandomTime() {
	t1, _ := time.Parse(time.RFC1123Z, "Mon, 22 Jun 2020 09:00:00 +0000")
	t2, _ := time.Parse(time.RFC1123Z, "Mon, 22 Jun 2020 09:45:00 +0000")

	randTime := randomTimeBetween(t1, t2)
	suite.Assert().Truef(!randTime.Before(t1), "randTime must be after %s but is %s ", t1, randTime)
	suite.Assert().Truef(randTime.Before(t2), "randTime must be before %s but is %s ", t2, randTime)
}
