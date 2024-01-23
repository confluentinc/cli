package ccloud

import (
	"net/http"

	"github.com/dghubble/sling"
)

const (
	signupEndpoint      = "/api/signup"
	verifyEmailEndpoint = "/api/email_verifications"
)

type SignupService struct {
	client *http.Client
	sling  *sling.Sling
}

func NewSignupService(client *Client) *SignupService {
	return &SignupService{
		client: client.HttpClient,
		sling:  client.sling,
	}
}

func (s *SignupService) Create(req *SignupRequest) (*SignupReply, error) {
	res := new(SignupReply)
	_, err := s.sling.New().Post(signupEndpoint).BodyProvider(Request(req)).Receive(res, res)
	return res, WrapErr(ReplyErr(res, err), "failed to sign up")
}

func (s *SignupService) SendVerificationEmail(user *User) error {
	req := CreateEmailVerificationRequest{
		Credentials: &Credentials{
			Username: user.Email,
		},
	}
	_, err := s.sling.New().Post(verifyEmailEndpoint).BodyProvider(Request(&req)).Receive(nil, nil)
	return WrapErr(err, "failed to send verification email")
}
