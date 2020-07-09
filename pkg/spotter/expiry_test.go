package spotter

import (
	"fmt"
	"time"

	"github.com/roppenlabs/silent-assassin/pkg/config"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ExpiryTestData struct {
	NodeName     string
	CreationTime time.Time
	EligibleWLs  []TimeSpan
}
type TimeSpan struct {
	Start time.Time
	End   time.Time
}

func parseTime(t string) time.Time {
	r, _ := time.Parse(time.RFC1123Z, t)
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
			NodeName:     "Node-1",
			CreationTime: parseTime("Mon, 22 Jun 2020 10:10:00 +0000"),
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
			NodeName:     "Node-2",
			CreationTime: parseTime("Mon, 22 Jun 2020 15:40:00 +0000"),

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
			NodeName:     "Node-3",
			CreationTime: parseTime("Mon, 22 Jun 2020 00:40:00 +0000"),
			EligibleWLs: []TimeSpan{
				{
					Start: parseTime("Mon, 22 Jun 2020 12:00:00 +0000"),
					End:   parseTime("Mon, 22 Jun 2020 14:00:00 +0000"),
				},
			},
		},
		{
			NodeName:     "Node-4",
			CreationTime: parseTime("Mon, 22 Jun 2020 22:20:00 +0000"),
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
	}

	suite.configMock.On("GetStringSlice", config.SpotterWhiteListIntervalHours).Return([]string{"00:00-06:00", "12:00-14:00"})
	ss := NewSpotterService(suite.configMock, suite.logger, suite.k8sMock)
	ss.initWhitelist()

	for _, testInput := range testData {

		creationTimestamp := testInput.CreationTime

		nodeToBeAnnotated := v1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name:              testInput.NodeName,
				CreationTimestamp: metav1.NewTime(creationTimestamp),
				Annotations:       map[string]string{"node.alpha.kubernetes.io/ttl": "0"}}}

		saExpTime, _ := time.Parse(time.RFC1123Z, ss.getExpiryTimestamp(nodeToBeAnnotated))

		assert.True(suite.T(), verifyNodeExpiry(saExpTime, testInput.EligibleWLs), fmt.Sprintf("SA_Expiry time =[ %v ] didn't fall within one of the eligible WL interval = [ %v ] for Node = %v", saExpTime, testInput.EligibleWLs, testInput.NodeName))
	}
}

func (suite *SpotterTestSuite) TestShouldReturnETinSameTimeZoneAsCT() {

	suite.configMock.On("GetStringSlice", config.SpotterWhiteListIntervalHours).Return([]string{"00:00-06:00", "12:00-14:00"})
	ss := NewSpotterService(suite.configMock, suite.logger, suite.k8sMock)
	ss.initWhitelist()
	creationTime := parseTime("Mon, 22 Jun 2020 22:20:00 +0530")
	nodeToBeAnnotated := v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "Node-IST",
			CreationTimestamp: metav1.NewTime(creationTime),
			Annotations:       map[string]string{"node.alpha.kubernetes.io/ttl": "0"}}}

	saExpTime, _ := time.Parse(time.RFC1123Z, ss.getExpiryTimestamp(nodeToBeAnnotated))

	assert.True(suite.T(), saExpTime.Location() == creationTime.Location(), "CT and ET TimeZone does not match")

}

func (suite *SpotterTestSuite) TestRandomNumber() {
	n1 := randomNumber(10, 30)
	n2 := randomNumber(10, 30)

	assert.True(suite.T(), n1 >= 10, fmt.Sprintf("Expected >10 got %d", n1))
	assert.True(suite.T(), n1 <= 30, fmt.Sprintf("Expected <30 got %d", n1))
	assert.True(suite.T(), n2 >= 10, fmt.Sprintf("Expected >10 got %d", n2))
	assert.True(suite.T(), n2 <= 30, fmt.Sprintf("Expected <30 got %d", n2))
}

func (suite *SpotterTestSuite) TestRandomMins() {
	t1, _ := time.Parse(time.RFC1123Z, "Mon, 22 Jun 2020 09:00:00 +0000")
	t2, _ := time.Parse(time.RFC1123Z, "Mon, 22 Jun 2020 09:45:00 +0000")

	randMins := randomMinuntes(t1, t2)
	assert.True(suite.T(), randMins >= 0*time.Minute, fmt.Sprintf("randMins must be >=0 but is: %d ", randMins))
	assert.True(suite.T(), randMins <= 45*time.Minute, fmt.Sprintf("randMins must be <= 45 but is: %d ", randMins))
}
