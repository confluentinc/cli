package common

import (
	"fmt"

	chttp "github.com/confluentinc/cli/http"
)

func HandleError(err error) error {
	switch err {
	case chttp.ErrUnauthorized:
		fmt.Println("You must login to access Confluent Cloud.")
	case chttp.ErrExpiredToken:
		fmt.Println("Your access to Confluent Cloud has expired. Please login again.")
	case chttp.ErrIncorrectAuth:
		fmt.Println("You have entered an incorrect username or password. Please try again.")
	default:
		return err
	}
	return nil
}
