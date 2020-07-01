package spotter

import (
	"time"

	"github.com/roppenlabs/silent-assassin/pkg/config"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (suite *SpotterTestSuite) TestShouldPanicIfTTLIsNegative() {
	creationTimestamp, _ := time.Parse(time.RFC1123Z, "Mon, 22 Jun 2020 12:54:45 +0530")
	TTL := -10
	nodeToBeAnnotated := v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "Node-2",
			CreationTimestamp: metav1.NewTime(creationTimestamp),
			Annotations:       map[string]string{"node.alpha.kubernetes.io/ttl": "0"}}}

	ss := NewSpotterService(suite.configMock, suite.logger, suite.k8sMock)

	assert.Panics(suite.T(), func() { ss.getExpiryTimestamp(nodeToBeAnnotated, TTL) }, "Spotter must panic if TTL is negative")

}

//If the CET (CT+TTL) falls in a WhiteList Interval , it should be used as is
func (suite *SpotterTestSuite) TestShouldNotUpdateCETWhenCETFallsInWLIntervals() {
	suite.configMock.On("GetStringSlice", config.SpotterWhiteListIntervalHours).Return([]string{"00:00-06:00", "12:00-14:00"})
	creationTimestamp, _ := time.Parse(time.RFC1123Z, "Mon, 22 Jun 2020 09:30:00 +0000")
	TTL := 15
	expectedExpiryTimestamp := creationTimestamp.Add(15 * time.Hour)
	nodeToBeAnnotated := v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "Node-2",
			CreationTimestamp: metav1.NewTime(creationTimestamp),
			Annotations:       map[string]string{"node.alpha.kubernetes.io/ttl": "0"}}}

	ss := NewSpotterService(suite.configMock, suite.logger, suite.k8sMock)
	ss.initWhitelist()

	assert.Equal(suite.T(), expectedExpiryTimestamp.Format(time.RFC1123Z), ss.getExpiryTimestamp(nodeToBeAnnotated, TTL), "Spotter must panic if TTL is negative")

}

//If the CET (CT+TTL) doesn't falls in a WhiteList Interval , it should pick from AdjustedET (incremented and decremented in step of 30 mins )
// that is closer to actual CET (CT+TTL) Incremented case
func (suite *SpotterTestSuite) TestShouldPickAdjustedETCloserToActualCETWhenCETDontFallInWLIntervalsDecrementCase() {
	suite.configMock.On("GetStringSlice", config.SpotterWhiteListIntervalHours).Return([]string{"00:00-06:00", "12:00-14:00"})
	creationTimestamp, _ := time.Parse(time.RFC1123Z, "Mon, 22 Jun 2020 15:45:00 +0000")
	TTL := 15
	expectedExpiryTimestamp, _ := time.Parse(time.RFC1123Z, "Mon, 23 Jun 2020 05:45:00 +0000")
	nodeToBeAnnotated := v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "Node-2",
			CreationTimestamp: metav1.NewTime(creationTimestamp),
			Annotations:       map[string]string{"node.alpha.kubernetes.io/ttl": "0"}}}

	ss := NewSpotterService(suite.configMock, suite.logger, suite.k8sMock)
	ss.initWhitelist()

	assert.Equal(suite.T(), expectedExpiryTimestamp.Format(time.RFC1123Z), ss.getExpiryTimestamp(nodeToBeAnnotated, TTL), "Spotter must panic if TTL is negative")

}

func (suite *SpotterTestSuite) TestShouldPickAdjustedETCloserToActualCETWhenCETDontFallInWLIntervalsIncrementCase() {
	suite.T().Skip()
	suite.configMock.On("GetStringSlice", config.SpotterWhiteListIntervalHours).Return([]string{"00:00-06:00", "12:00-14:00"})
	creationTimestamp, _ := time.Parse(time.RFC1123Z, "Mon, 22 Jun 2020 09:45:00 +0000")
	TTL := 14
	expectedExpiryTimestamp, _ := time.Parse(time.RFC1123Z, "Mon, 22 Jun 2020 00:15:00 +0000")
	nodeToBeAnnotated := v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "Node-2",
			CreationTimestamp: metav1.NewTime(creationTimestamp),
			Annotations:       map[string]string{"node.alpha.kubernetes.io/ttl": "0"}}}

	ss := NewSpotterService(suite.configMock, suite.logger, suite.k8sMock)
	ss.initWhitelist()

	assert.Equal(suite.T(), expectedExpiryTimestamp.Format(time.RFC1123Z), ss.getExpiryTimestamp(nodeToBeAnnotated, TTL), "Spotter must panic if TTL is negative")

}

func (suite *SpotterTestSuite) TestShouldAdd24HrsIfCETIsBeforeCreationTime() {
	suite.configMock.On("GetStringSlice", config.SpotterWhiteListIntervalHours).Return([]string{"08:00-09:00", "03:00-05:00"})
	creationTimestamp, _ := time.Parse(time.RFC1123Z, "Mon, 22 Jun 2020 09:45:00 +0000")
	TTL := 6
	expectedExpiryTimestamp, _ := time.Parse(time.RFC1123Z, "Mon, 23 Jun 2020 08:45:00 +0000")
	nodeToBeAnnotated := v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "Node-2",
			CreationTimestamp: metav1.NewTime(creationTimestamp),
			Annotations:       map[string]string{"node.alpha.kubernetes.io/ttl": "0"}}}

	ss := NewSpotterService(suite.configMock, suite.logger, suite.k8sMock)
	ss.initWhitelist()

	assert.Equal(suite.T(), expectedExpiryTimestamp.Format(time.RFC1123Z), ss.getExpiryTimestamp(nodeToBeAnnotated, TTL), "Spotter must panic if TTL is negative")

}
