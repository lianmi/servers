package auth

import (
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"strconv"
)

// the questions to ask
var qs = []*survey.Question{
	{
		Name:      "name",
		Prompt:    &survey.Input{Message: "What is your name?"},
		Validate:  survey.Required,
		Transform: survey.Title,
	},
	{
		Name: "color",
		Prompt: &survey.Select{
			Message: "Choose a color:",
			Options: []string{"red", "blue", "green"},
			Default: "red",
		},
	},
	{
		Name: "age",
		Validate: func(val interface{}) error {
			// if the input matches the expectation
			if val == "" {
				return fmt.Errorf("Age must required")
			}
			//int, err := strconv.Atoi(string) //string转成int
			//int64, err := strconv.ParseInt(string, 10, 64) //string转成int64
			if age, err := strconv.Atoi(val.(string)); err != nil {
				return fmt.Errorf("You entered a number: %s", val.(string))
			} else {
				if age < 0 {
					return fmt.Errorf("Age:%d must gather than 0", age)
				}
			}
			// nothing was wrong
			return nil
		},
		Prompt: &survey.Input{Message: "How old are you?"},
	},
}
