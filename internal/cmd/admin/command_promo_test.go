package admin

import (
	"context"
	"testing"
	"time"

	ccloudv1 "github.com/confluentinc/ccloud-sdk-go-v1-public"
	ccloudv1mock "github.com/confluentinc/ccloud-sdk-go-v1-public/mock"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	climock "github.com/confluentinc/cli/mock"
	"github.com/stretchr/testify/require"
)

func TestPromoAdd(t *testing.T) {
	client := &ccloudv1.Client{
		Billing: &ccloudv1mock.Billing{
			ClaimPromoCodeFunc: func(_ context.Context, _ *ccloudv1.Organization, _ string) (*ccloudv1.PromoCodeClaim, error) {
				return nil, nil
			},
		},
	}

	cfg := v1.AuthenticatedCloudConfigMock()
	cmd := New(climock.NewPreRunnerMock(client, nil, nil, nil, cfg), true)

	out, err := pcmd.ExecuteCommand(cmd, "promo", "add", "XXXXX")
	require.NoError(t, err)
	require.Equal(t, "Your promo code was successfully added.\n", out)
}

func TestPromoListEmpty(t *testing.T) {
	client := &ccloudv1.Client{
		Billing: &ccloudv1mock.Billing{
			GetClaimedPromoCodesFunc: func(_ context.Context, _ *ccloudv1.Organization, _ bool) ([]*ccloudv1.PromoCodeClaim, error) {
				return []*ccloudv1.PromoCodeClaim{}, nil
			},
		},
	}

	cfg := v1.AuthenticatedCloudConfigMock()
	cmd := New(climock.NewPreRunnerMock(client, nil, nil, nil, cfg), true)

	out, err := pcmd.ExecuteCommand(cmd, "promo", "list")
	require.NoError(t, err)
	require.Equal(t, "None found.\n", out)
}

func TestFormatBalance(t *testing.T) {
	require.Equal(t, "$0.00/1.00 USD", formatBalance(0, 10000))
}

func TestConvertToUSD(t *testing.T) {
	require.Equal(t, 1.23, ConvertToUSD(12300))
}

func TestFormatExpiration(t *testing.T) {
	date := time.Date(2021, time.June, 16, 0, 0, 0, 0, time.Local)
	require.Equal(t, "Jun 16, 2021", formatExpiration(date.Unix()))
}
