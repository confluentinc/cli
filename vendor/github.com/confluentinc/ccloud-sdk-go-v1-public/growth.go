package ccloud

import (
	"net/http"

	"github.com/dghubble/sling"
)

const (
	freeTrialInfoEndpoint = "/api/growth/v1/free-trial-info"
)

type GrowthService struct {
	client *http.Client
	sling  *sling.Sling
}

func NewGrowthService(client *Client) *GrowthService {
	return &GrowthService{
		client: client.HttpClient,
		sling:  client.sling,
	}
}

func (o *GrowthService) GetFreeTrialInfo(orgId int32) ([]*GrowthPromoCodeClaim, error) {
	req := &GetFreeTrialInfoRequest{OrgId: orgId}
	res := &GetFreeTrialInfoReply{}

	_, err := o.sling.New().Get(freeTrialInfoEndpoint).BodyProvider(Request(req)).Receive(res, res)
	return res.PromoCodeClaims, WrapErr(ReplyErr(res, err), "failed to get free trial info")
}
