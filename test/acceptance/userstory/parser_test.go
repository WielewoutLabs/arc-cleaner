package userstory_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/wielewout/arc-cleaner/test/acceptance/userstory"
)

type ParseTestSuite struct {
	suite.Suite
}

func (suite *ParseTestSuite) TestShouldParsePersona() {
	userStory, err := userstory.Parse("As a type of user, I want to perform some task so that I can achieve some goal.")
	suite.Require().NoError(err)

	suite.Require().Equal("a type of user", userStory.Persona)
}

func (suite *ParseTestSuite) TestShouldParseIntent() {
	userStory, err := userstory.Parse("As a type of user, I want to perform some task so that I can achieve some goal.")
	suite.Require().NoError(err)

	suite.Require().Equal("perform some task", userStory.Intent)
}

func (suite *ParseTestSuite) TestShouldParseReason() {
	userStory, err := userstory.Parse("As a type of user, I want to perform some task so that I can achieve some goal.")
	suite.Require().NoError(err)

	suite.Require().Equal("I can achieve some goal", userStory.Reason)
}

func (suite *ParseTestSuite) TestShouldParseUserStoryWithSameString() {
	userStory, err := userstory.Parse("As a type of user, I want to perform some task so that I can achieve some goal.")
	suite.Require().NoError(err)

	suite.Require().Equal("As a type of user, I want to perform some task so that I can achieve some goal.", userStory.String())
}

func (suite *ParseTestSuite) TestShouldIgnoreWhitespaceBeforeAs() {
	userStory, err := userstory.Parse("   As a type of user, I want to perform some task so that I can achieve some goal.")
	suite.Require().NoError(err)

	suite.Require().Equal("As a type of user, I want to perform some task so that I can achieve some goal.", userStory.String())
}

func (suite *ParseTestSuite) TestShouldIgnoreWhitespaceAfterEnd() {
	userStory, err := userstory.Parse("As a type of user, I want to perform some task so that I can achieve some goal.\n    ")
	suite.Require().NoError(err)

	suite.Require().Equal("As a type of user, I want to perform some task so that I can achieve some goal.", userStory.String())
}

func (suite *ParseTestSuite) TestShouldErrorWhenNotStartingWithAs() {
	_, err := userstory.Parse("For a type of user, I want to perform some task so that I can achieve some goal.")

	suite.Require().Error(err)
	suite.Require().Equal(`story is not starting with "As"`, err.Error())
}

func (suite *ParseTestSuite) TestShouldErrorWhenAsNotEndedWithComma() {
	_, err := userstory.Parse("As a type of user I want to perform some task so that I can achieve some goal.")

	suite.Require().Error(err)
	suite.Require().Equal(`story does not contain an ending for "As" (indicated with a comma ",")`, err.Error())
}

func (suite *ParseTestSuite) TestShouldErrorWhenNotContainingIWantTo() {
	_, err := userstory.Parse("As a type of user, so that I can achieve some goal.")

	suite.Require().Error(err)
	suite.Require().Equal(`story does not contain "I want to"`, err.Error())
}

func (suite *ParseTestSuite) TestShouldErrorWhenAsNotImmediatelyFollowedByIWantTo() {
	_, err := userstory.Parse("As a type of user, sometimes I want to perform some task so that I can achieve some goal.")

	suite.Require().Error(err)
	suite.Require().Equal(`story does not immediately follow up the ending for "As" (indicated with a comma ",") with "I want to"`, err.Error())
}

func (suite *ParseTestSuite) TestShouldErrorWhenNotContainingSoThat() {
	_, err := userstory.Parse("As a type of user, I want to perform some task.")

	suite.Require().Error(err)
	suite.Require().Equal(`story does not contain "so that"`, err.Error())
}

func (suite *ParseTestSuite) TestShouldErrorWhenUserStoryNotEnded() {
	_, err := userstory.Parse("As a type of user, I want to perform some task so that I can achieve some goal")

	suite.Require().Error(err)
	suite.Require().Equal(`story does not end with "."`, err.Error())
}

func (suite *ParseTestSuite) TestShouldErrorWhenEndingPresentInAs() {
	_, err := userstory.Parse("As a type. of user, I want to perform some task so that I can achieve some goal.")

	suite.Require().Error(err)
	suite.Require().Equal(`story indicates an ending in "As"`, err.Error())
}

func (suite *ParseTestSuite) TestShouldErrorWhenEndingPresentInIWantTo() {
	_, err := userstory.Parse("As a type of user, I want to perform. some task so that I can achieve some goal.")

	suite.Require().Error(err)
	suite.Require().Equal(`story indicates an ending in "I want to"`, err.Error())
}

func TestParse(t *testing.T) {
	suite.Run(t, new(ParseTestSuite))
}
