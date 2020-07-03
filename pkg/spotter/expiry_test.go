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
	NodeName      string
	CreationTime  time.Time
	TargetWLStart time.Time
	TargetWLEnd   time.Time
	WL            []string
}

func parseTime(t string) time.Time {
	r, _ := time.Parse(time.RFC1123Z, t)
	return r
}

//If the CET (CT+TTL) falls in a WhiteList Interval , it should be used as is
func (suite *SpotterTestSuite) TestShouldNotUpdateCETWhenCETFallsInWLIntervals() {
	testData := []ExpiryTestData{
		{
			NodeName:      "Node-1",
			CreationTime:  parseTime("Mon, 22 Jun 2020 10:10:00 +0000"),
			TargetWLStart: parseTime("Mon, 23 Jun 2020 00:00:00 +0000"),
			TargetWLEnd:   parseTime("Mon, 23 Jun 2020 06:00:00 +0000"),
			WL:            []string{"00:00-06:00", "12:00-14:00"},
		},
		{
			NodeName:      "Node-2",
			CreationTime:  parseTime("Mon, 22 Jun 2020 15:40:00 +0000"),
			TargetWLStart: parseTime("Mon, 23 Jun 2020 12:00:00 +0000"),
			TargetWLEnd:   parseTime("Mon, 23 Jun 2020 14:00:00 +0000"),
			WL:            []string{"00:00-06:00", "12:00-14:00"},
		},
		{
			NodeName:      "Node-3",
			CreationTime:  parseTime("Mon, 22 Jun 2020 00:40:00 +0000"),
			TargetWLStart: parseTime("Mon, 22 Jun 2020 12:00:00 +0000"),
			TargetWLEnd:   parseTime("Mon, 22 Jun 2020 14:00:00 +0000"),
			WL:            []string{"00:00-06:00", "12:00-14:00"},
		},
		{
			NodeName:      "Node-4",
			CreationTime:  parseTime("Mon, 22 Jun 2020 22:20:00 +0000"),
			TargetWLStart: parseTime("Mon, 23 Jun 2020 12:00:00 +0000"),
			TargetWLEnd:   parseTime("Mon, 23 Jun 2020 14:00:00 +0000"),
			WL:            []string{"00:00-06:00", "12:00-14:00"},
		},
	}
	for _, testInput := range testData {
		suite.configMock.On("GetStringSlice", config.SpotterWhiteListIntervalHours).Return(testInput.WL)
		creationTimestamp := testInput.CreationTime

		targetWLStart := testInput.TargetWLStart
		targetWLEnd := testInput.TargetWLEnd

		nodeToBeAnnotated := v1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name:              "Node-2",
				CreationTimestamp: metav1.NewTime(creationTimestamp),
				Annotations:       map[string]string{"node.alpha.kubernetes.io/ttl": "0"}}}

		ss := NewSpotterService(suite.configMock, suite.logger, suite.k8sMock)
		ss.initWhitelist()
		saExpTime, _ := time.Parse(time.RFC1123Z, ss.getExpiryTimestamp(nodeToBeAnnotated))

		assert.True(suite.T(), saExpTime.After(targetWLStart) || saExpTime.Equal(targetWLStart), fmt.Sprintf("SA_Expiry time =[ %v ] must be After or Equal to the Start of target WL interval = [ %v ] for Node = %v", saExpTime, targetWLStart, testInput.NodeName))
		assert.True(suite.T(), saExpTime.Before(targetWLEnd) || saExpTime.Equal(targetWLEnd), fmt.Sprintf("SA_Expiry time =[ %v ] must be Before or Equal to the End of target WL interval = [ %v ] for Node = %v", saExpTime, targetWLEnd, testInput.NodeName))
	}
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
	assert.True(suite.T(), randMins > 0*time.Minute)
	assert.True(suite.T(), randMins < 45*time.Minute)
}
