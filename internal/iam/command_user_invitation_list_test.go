package iam

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	iamv2 "github.com/confluentinc/ccloud-sdk-go-v2/iam/v2"
)

func TestUserNames(t *testing.T) {
	names, err := userNames(func() ([]iamv2.IamV2User, error) {
		return []iamv2.IamV2User{
			{Id: iamv2.PtrString("u-111"), FullName: iamv2.PtrString("Ada Lovelace")},
			{Id: iamv2.PtrString("u-222"), FullName: iamv2.PtrString("Grace Hopper")},
		}, nil
	})
	require.NoError(t, err)
	require.Equal(t, map[string]string{"u-111": "Ada Lovelace", "u-222": "Grace Hopper"}, names)
	require.Empty(t, names["u-missing"])
}

func TestUserNamesListFailure(t *testing.T) {
	names, err := userNames(func() ([]iamv2.IamV2User, error) {
		return nil, errors.New("boom")
	})
	require.Error(t, err)
	require.NotNil(t, names)
	require.Empty(t, names["u-111"])
}
