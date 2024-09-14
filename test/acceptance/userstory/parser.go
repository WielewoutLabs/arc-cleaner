package userstory

import (
	"errors"
	"strings"
)

const (
	as             = "As "
	endOfAs        = ", "
	iWantTo        = "I want to "
	soThat         = " so that "
	endOfUserStory = "."
)

func Parse(rawStory string) (UserStory, error) {
	trimmedRawStory := strings.TrimSpace(rawStory)

	asIndex := strings.Index(trimmedRawStory, as)
	if asIndex != 0 {
		return UserStory{}, errors.New(`story is not starting with "As"`)
	}

	endOfAsAndIWantToIndex := strings.Index(trimmedRawStory, endOfAs+iWantTo)
	if endOfAsAndIWantToIndex == -1 {
		endOfAsIndex := strings.Index(trimmedRawStory, endOfAs)
		if endOfAsIndex == -1 {
			return UserStory{}, errors.New(`story does not contain an ending for "As" (indicated with a comma ",")`)
		}

		iWantToIndex := strings.Index(trimmedRawStory, iWantTo)
		if iWantToIndex == -1 {
			return UserStory{}, errors.New(`story does not contain "I want to"`)
		}

		return UserStory{}, errors.New(`story does not immediately follow up the ending for "As" (indicated with a comma ",") with "I want to"`)
	}

	soThatIndex := strings.Index(trimmedRawStory, soThat)
	if soThatIndex == -1 {
		return UserStory{}, errors.New(`story does not contain "so that"`)
	}

	endOfUserStoryIndex := strings.Index(trimmedRawStory, endOfUserStory)
	if endOfUserStoryIndex == -1 {
		return UserStory{}, errors.New(`story does not end with "."`)
	}

	if endOfUserStoryIndex < endOfAsAndIWantToIndex {
		return UserStory{}, errors.New(`story indicates an ending in "As"`)
	}
	if endOfUserStoryIndex < soThatIndex {
		return UserStory{}, errors.New(`story indicates an ending in "I want to"`)
	}

	persona := trimmedRawStory[asIndex+len(as) : endOfAsAndIWantToIndex]
	intent := trimmedRawStory[endOfAsAndIWantToIndex+len(endOfAs+iWantTo) : soThatIndex]
	reason := trimmedRawStory[soThatIndex+len(soThat) : endOfUserStoryIndex]

	return UserStory{
		Persona: persona,
		Intent:  intent,
		Reason:  reason,
	}, nil
}
