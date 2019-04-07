package s3

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/confluentinc/cli/internal/pkg/update/mock"
	"github.com/stretchr/testify/require"
)

func TestNewPrivateRepo(t *testing.T) {
	badCreds := makeCreds("", credentials.Value{}, fmt.Errorf("oops"), false)
	goodCreds := makeCreds("", credentials.Value{AccessKeyID: "ak"}, nil, false)
	tests := []struct {
		name    string
		params  *PrivateRepoParams
		want    *PrivateRepo
		wantErr bool
	}{
		{
			name: "should error if region not provided",
			params: &PrivateRepoParams{
				S3BinBucket: "bucket",
				S3BinRegion: "",
				S3BinPrefix: "prefix",
				S3ObjectKey: &mock.ObjectKey{},
			},
			wantErr: true,
		},
		{
			name: "should error if bucket not provided",
			params: &PrivateRepoParams{
				S3BinBucket: "",
				S3BinRegion: "region",
				S3BinPrefix: "prefix",
				S3ObjectKey: &mock.ObjectKey{},
			},
			wantErr: true,
		},
		{
			name: "will error if empty prefix (TODO)",
			params: &PrivateRepoParams{
				S3BinBucket: "bucket",
				S3BinRegion: "region",
				S3BinPrefix: "",
				S3ObjectKey: &mock.ObjectKey{},
			},
			wantErr: true,
		},
		{
			name: "should error if invalid credentials",
			params: &PrivateRepoParams{
				S3BinBucket: "bucket",
				S3BinRegion: "region",
				S3BinPrefix: "prefix",
				S3ObjectKey: &mock.ObjectKey{},
				creds:       badCreds,
			},
			wantErr: true,
		},
		{
			name: "should return private pkg repo",
			params: &PrivateRepoParams{
				S3BinBucket:  "bucket",
				S3BinRegion:  "region",
				S3BinPrefix:  "prefix",
				S3ObjectKey:  &mock.ObjectKey{},
				creds:        goodCreds,
				s3svc:        &mockS3Client{},
				s3downloader: &mockS3Downloader{},
			},
			want: &PrivateRepo{
				PrivateRepoParams: &PrivateRepoParams{
					S3BinBucket:  "bucket",
					S3BinRegion:  "region",
					S3BinPrefix:  "prefix",
					S3ObjectKey:  &mock.ObjectKey{},
					creds:        goodCreds,
					s3svc:        &mockS3Client{},
					s3downloader: &mockS3Downloader{},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewPrivateRepo(tt.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewPrivateRepo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewPrivateRepo() = %#v, want %#v\n params = %#v, want params =%#v",
					got, tt.want, got.PrivateRepoParams, tt.want.PrivateRepoParams)
			}
		})
	}
}

func Test_getCredentials(t *testing.T) {
	type args struct {
		envVar      string
		cf          *mockCredsFactory
		allProfiles []string
	}
	tests := []struct {
		name       string
		args       args
		want       credentials.Value
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "should use default profile if no profiles given and no AWS_PROFILE set",
			args: args{
				cf: makeCreds("default", credentials.Value{AccessKeyID: "ak"}, nil, false),
			},
			want: credentials.Value{
				AccessKeyID: "ak",
			},
		},
		{
			name: "should check AWS_PROFILE env var if no profiles given",
			args: args{
				envVar: "my-other-profile",
				cf:     makeCreds("my-other-profile", credentials.Value{AccessKeyID: "ak"}, nil, false),
			},
			want: credentials.Value{
				AccessKeyID: "ak",
			},
		},
		{
			name: "should error if access key id is empty",
			args: args{
				cf: makeCreds("default", credentials.Value{}, nil, false),
			},
			wantErr: true,
		},
		{
			name: "should error if credentials are expired",
			args: args{
				cf: makeCreds("default", credentials.Value{AccessKeyID: "ak"}, nil, true),
			},
			wantErr: true,
		},
		{
			name: "should search multiple profiles if given",
			args: args{
				allProfiles: []string{"profile1", "profile2", "profile3"},
				cf: &mockCredsFactory{allCreds: []credsAssert{
					{expectProfile: "profile1", provider: &mockCredentialsProvider{err: fmt.Errorf("error1")}},
					{expectProfile: "profile2", provider: &mockCredentialsProvider{err: fmt.Errorf("error2")}},
					{expectProfile: "profile3", provider: &mockCredentialsProvider{val: credentials.Value{AccessKeyID: "VAULT"}}},
				}},
			},
			want: credentials.Value{AccessKeyID: "VAULT"},
		},
		{
			name: "should search multiple profiles if given, with env var as final fallback",
			args: args{
				envVar:      "profile4",
				allProfiles: []string{"profile1", "profile2", "profile3"},
				cf: &mockCredsFactory{allCreds: []credsAssert{
					{expectProfile: "profile1", provider: &mockCredentialsProvider{err: fmt.Errorf("error1")}},
					{expectProfile: "profile2", provider: &mockCredentialsProvider{err: fmt.Errorf("error2")}},
					{expectProfile: "profile3", provider: &mockCredentialsProvider{err: fmt.Errorf("error3")}},
					{expectProfile: "profile4", provider: &mockCredentialsProvider{val: credentials.Value{AccessKeyID: "VAULT"}}},
				}},
			},
			want: credentials.Value{AccessKeyID: "VAULT"},
		},
		{
			name: "should reformat errors to be more easily readable - single profile",
			args: args{
				cf: makeCreds("default", credentials.Value{}, nil, true),
			},
			wantErr: true,
			wantErrMsg: `2 errors occurred:
	* failed to find aws credentials in profiles: default
	*   error: access key id is empty for default

`,
		},
		{
			name: "should reformat errors to be more easily readable - multiple profiles",
			args: args{
				envVar:      "profile4",
				allProfiles: []string{"profile1", "profile2", "profile3"},
				cf: &mockCredsFactory{allCreds: []credsAssert{
					{expectProfile: "profile1", provider: &mockCredentialsProvider{expired: true, val: credentials.Value{AccessKeyID: "VAULT"}}},
					{expectProfile: "profile2", provider: &mockCredentialsProvider{val: credentials.Value{}}},
					{expectProfile: "profile3", provider: &mockCredentialsProvider{err: fmt.Errorf("error3")}},
					{expectProfile: "profile4", provider: &mockCredentialsProvider{err: fmt.Errorf("error4")}},
				}},
			},
			wantErr: true,
			wantErrMsg: `5 errors occurred:
	* failed to find aws credentials in profiles: profile1, profile2, profile3, profile4
	*   error: aws creds in profile profile1 are expired
	*   error: access key id is empty for profile2
	*   error while finding creds: error3
	*   error while finding creds: error4

`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := require.New(t)
			if tt.args.envVar != "" {
				oldEnv, found := os.LookupEnv("AWS_PROFILE")
				req.NoError(os.Setenv("AWS_PROFILE", tt.args.envVar))
				defer func() {
					if found {
						req.NoError(os.Setenv("AWS_PROFILE", oldEnv))
					} else {
						req.NoError(os.Unsetenv("AWS_PROFILE"))
					}
				}()
			}

			tt.args.cf.req = req
			got, err := getCredentials(tt.args.cf, tt.args.allProfiles)
			if (err != nil) != tt.wantErr {
				t.Errorf("getCredentials() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErrMsg != "" {
				req.Equal(tt.wantErrMsg, err.Error())
			}
			if err != nil {
				req.Nil(got)
				return
			}
			creds, err := got.Get()
			req.NoError(err)
			if !reflect.DeepEqual(creds, tt.want) {
				t.Errorf("getCredentials() = %#v, want %#v", creds, tt.want)
			}
		})
	}
}

type mockCredentialsProvider struct {
	val     credentials.Value
	err     error
	expired bool
}

func (m *mockCredentialsProvider) Retrieve() (credentials.Value, error) {
	return m.val, m.err
}

func (m *mockCredentialsProvider) IsExpired() bool { return m.expired }

type credsAssert struct {
	provider      credentials.Provider
	expectProfile string
}

type mockCredsFactory struct {
	allCreds []credsAssert
	req      *require.Assertions
	count    int
}

func (m *mockCredsFactory) newProvider(profile string) credentials.Provider {
	creds := m.allCreds[m.count]
	if creds.expectProfile != "" && m.req != nil {
		m.req.Equal(creds.expectProfile, profile)
	}
	m.count++
	return creds.provider
}

type mockS3Client struct {
	s3iface.S3API
}

type mockS3Downloader struct{}

func (d *mockS3Downloader) Download(w io.WriterAt, input *s3.GetObjectInput, options ...func(*s3manager.Downloader)) (n int64, err error) {
	return 0, nil
}

func makeCreds(profile string, val credentials.Value, err error, expired bool) *mockCredsFactory {
	return &mockCredsFactory{allCreds: []credsAssert{
		{expectProfile: profile, provider: &mockCredentialsProvider{val, err, expired}},
	}}
}
