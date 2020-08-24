package notifier

import (
	"testing"

	"github.com/roppenlabs/silent-assassin/pkg/config"
	"github.com/roppenlabs/silent-assassin/pkg/logger"
	"github.com/roppenlabs/silent-assassin/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type notifierTestSuit struct {
	suite.Suite
	configMock *config.ProviderMock
	logger     logger.IZapLogger
	httpMock   *utils.MockHTTPClient
}

func (suite *notifierTestSuit) SetupTest() {
	suite.configMock = new(config.ProviderMock)
	suite.configMock.On("GetString", config.LogLevel).Return("debug")
	suite.logger = logger.Init(suite.configMock)
}

func (suite *notifierTestSuit) TestSouldInitializeSlackClient() {
	suite.configMock.On("GetString", config.SlackWebhookURL).Return("https://hooks.slack.com/services/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")
	suite.configMock.On("GetString", config.SlackUsername).Return("silent-assassin")
	suite.configMock.On("GetString", config.SlackChannel).Return("silent-assaain-dev")
	suite.configMock.On("GetString", config.SlackIconURL).Return("https://www.flaticon.com/free-icon/slack_2111615")
	suite.configMock.On("GetUint32", config.SlackTimeoutMs).Return(uint32(5000))

	n := NewNotificationService(suite.configMock, suite.logger)

	assert.IsType(suite.T(), Slack{}, n.provider)
}

func TestNotifierTestSuite(t *testing.T) {
	suite.Run(t, new(notifierTestSuit))
}
