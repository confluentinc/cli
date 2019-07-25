package errors

import (
	"fmt"

	"github.com/pkg/errors"
)

/*
 * Invariants:
 * - Confluent SDK (http package) always returns a corev1.Error.
 * - Pkg always return an HTTP Error constant (top of this file)
 *
 * Error Flow:
 * - API error responses (json) are parsed into corev1.Error objects.
 *   - Note: API returns 404s for unauthorized resources, so HTTP package has to remap 404 -> 401 where appropriate.
 *
 * Create a custom error object if you need a custom field in your message (like the clusterID).
 * Otherwise just add a named error var.
 */

var (
	ErrNotImplemented = fmt.Errorf("not implemented")
	ErrNotLoggedIn    = fmt.Errorf("not logged in")
	ErrNoContext      = fmt.Errorf("context not set")
	ErrNoKafkaContext = fmt.Errorf("kafka not set")
	ErrNoSrEnabled    = fmt.Errorf("schema registry not enabled")
)

// UnspecifiedKafkaClusterError means the user needs to specify a kafka cluster
type UnspecifiedKafkaClusterError struct {
	KafkaClusterID string
}

func (e *UnspecifiedKafkaClusterError) Error() string {
	return e.KafkaClusterID
}

// UnspecifiedAPIKeyError means the user needs to set an api-key for this cluster
type UnspecifiedAPIKeyError struct {
	ClusterID string
}

func (e *UnspecifiedAPIKeyError) Error() string {
	return e.ClusterID
}

// UnconfiguredAPISecretError means the user needs to store the API secret locally
type UnconfiguredAPISecretError struct {
	APIKey    string
	ClusterID string
}

func (e *UnconfiguredAPISecretError) Error() string {
	return fmt.Sprintf("please add API secret with 'api-key store %s --cluster %s'", e.APIKey, e.ClusterID)
}

func New(msg string) error {
	return errors.New(msg)
}

func Wrap(err error, msg string) error {
	return errors.Wrap(err, msg)
}

func Wrapf(err error, fmt string, args ...interface{}) error {
	return errors.Wrapf(err, fmt, args...)
}

func Errorf(fmt string, args ...interface{}) error {
	return errors.Errorf(fmt, args...)
}

func Cause(err error) error {
	return errors.Cause(err)
}
