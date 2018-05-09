package common

import (
	"fmt"

	"github.com/confluentinc/cli/shared"
)

func HandleError(err error) error {
	switch err {
	case shared.ErrUnauthorized:
		fmt.Println("You must login to access Confluent Cloud.")
	case shared.ErrExpiredToken:
		fmt.Println("Your access to Confluent Cloud has expired. Please login again.")
	case shared.ErrIncorrectAuth:
		fmt.Println("You have entered an incorrect username or password. Please try again.")
	case shared.ErrMalformedToken:
		fmt.Println("Your auth token has been corrupted. Please login again.")
	case shared.ErrNotImplemented:
		fmt.Println("Sorry, this functionality is not yet available in the CLI.")
	case shared.ErrNotFound:
		fmt.Println("Kafka cluster not found.")  // TODO: parametrize ErrNotFound for better error messaging
	default:
		return err
	}
	return nil
}
