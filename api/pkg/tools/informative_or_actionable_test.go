package tools

import (
	"context"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/helixml/helix/api/pkg/types"
	"github.com/kelseyhightower/envconfig"
	"github.com/stretchr/testify/suite"
)

func TestActionTestSuite(t *testing.T) {
	suite.Run(t, new(ActionTestSuite))
}

type ActionTestSuite struct {
	suite.Suite
	ctx      context.Context
	strategy *ChainStrategy
}

func (suite *ActionTestSuite) SetupTest() {
	suite.ctx = context.Background()

	var cfg Config
	err := envconfig.Process("", &cfg)
	suite.NoError(err)

	spew.Dump(cfg)

	strategy, err := NewChainStrategy(&cfg)
	suite.NoError(err)

	suite.strategy = strategy
}

func (suite *ActionTestSuite) TestIsActionable() {
	tools := []*types.Tool{
		{
			Name:        "weatherAPI",
			Description: "Weather API that can return the current weather for a given location",
		},
		{
			Name:        "productsAPI",
			Description: "database API that can be used to query products information in the database",
		},
	}

	history := []*types.Interaction{}

	currentMessage := "What is the weather like in San Francisco?"

	resp, err := suite.strategy.IsActionable(suite.ctx, tools, history, currentMessage)
	suite.NoError(err)

	suite.Equal("yes", resp.NeedsApi)
	suite.Equal("weatherAPI", resp.Api)
}
