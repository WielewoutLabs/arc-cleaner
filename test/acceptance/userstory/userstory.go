package userstory

import (
	"fmt"
)

type UserStory struct {
	Persona string
	Intent  string
	Reason  string
}

func (us *UserStory) As(persona string) {
	us.Persona = persona
}

func (us *UserStory) IWantTo(intent string) {
	us.Intent = intent
}

func (us *UserStory) SoThat(reason string) {
	us.Reason = reason
}

func (us *UserStory) String() string {
	return fmt.Sprintf("As %s, I want to %s so that %s.", us.Persona, us.Intent, us.Reason)
}
