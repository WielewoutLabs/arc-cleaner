package acceptance_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/cucumber/godog"

	"github.com/wielewout/arc-cleaner/test/acceptance/steps"
	"github.com/wielewout/arc-cleaner/test/acceptance/system"
	"github.com/wielewout/arc-cleaner/test/acceptance/userstory"
)

func TestAcceptanceUserStoryValidation(t *testing.T) {
	systemConfig := system.NewConfig(t)

	suite := newGodogSuite(t, systemConfig)

	if err := validateUserStories(suite); err != nil {
		t.Fatalf("user story validation failed:\n%s", err.Error())
	}
}

func TestAcceptanceFeatures(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping acceptance feature tests")
	}

	systemConfig := system.SetUp(t)

	suite := newGodogSuite(t, systemConfig)

	if suite.Run() != 0 {
		t.Fatalf("non-zero status returned, failed to run feature tests")
	}
}

func newGodogSuite(t *testing.T, systemConfig system.Config) godog.TestSuite {
	suite := godog.TestSuite{
		ScenarioInitializer: func(ctx *godog.ScenarioContext) {
			initializeScenarios(ctx, systemConfig)
		},
		Options: &godog.Options{
			TestingT: t,
			Strict:   true,
			Format:   "pretty",
			Paths:    []string{"features"},
		},
	}

	return suite
}

func initializeScenarios(ctx *godog.ScenarioContext, systemConfig system.Config) {
	steps.InitializeHealthScenario(ctx, systemConfig)
}

func validateUserStories(suite godog.TestSuite) error {
	features, err := suite.RetrieveFeatures()
	if err != nil {
		return err
	}

	var joinedErrors error
	for _, f := range features {
		name := f.Feature.Name
		description := f.Feature.Description

		descriptionLines := strings.Split(description, "\n")
		trimmedDescriptionLines := make([]string, 0)
		for _, line := range descriptionLines {
			trimmedLine := strings.TrimSpace(line)
			trimmedDescriptionLines = append(trimmedDescriptionLines, trimmedLine)
		}
		trimmedDescription := strings.Join(trimmedDescriptionLines, " ")

		userStory, err := userstory.Parse(trimmedDescription)
		if err != nil {
			err = fmt.Errorf(`"%s" %w`, name, err)
			if joinedErrors == nil {
				joinedErrors = err
			} else {
				joinedErrors = errors.Join(joinedErrors, err)
			}
		}

		_ = userStory
	}
	if joinedErrors != nil {
		return joinedErrors
	}

	return nil
}
