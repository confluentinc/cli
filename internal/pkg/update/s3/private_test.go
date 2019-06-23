package s3

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/hashicorp/go-version"
	"github.com/stretchr/testify/require"

	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/log"
	pio "github.com/confluentinc/cli/internal/pkg/io"
	mock "github.com/confluentinc/cli/internal/pkg/mock"
	updateMock "github.com/confluentinc/cli/internal/pkg/update/mock"
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
				S3ObjectKey: &updateMock.ObjectKey{},
			},
			wantErr: true,
		},
		{
			name: "should error if bucket not provided",
			params: &PrivateRepoParams{
				S3BinBucket: "",
				S3BinRegion: "region",
				S3BinPrefix: "prefix",
				S3ObjectKey: &updateMock.ObjectKey{},
			},
			wantErr: true,
		},
		{
			name: "will error if empty prefix (TODO)",
			params: &PrivateRepoParams{
				S3BinBucket: "bucket",
				S3BinRegion: "region",
				S3BinPrefix: "",
				S3ObjectKey: &updateMock.ObjectKey{},
			},
			wantErr: true,
		},
		{
			name: "should error if invalid credentials",
			params: &PrivateRepoParams{
				S3BinBucket: "bucket",
				S3BinRegion: "region",
				S3BinPrefix: "prefix",
				S3ObjectKey: &updateMock.ObjectKey{},
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
				S3ObjectKey:  &updateMock.ObjectKey{},
				creds:        goodCreds,
				s3svc:        &updateMock.S3API{},
				s3downloader: &updateMock.Downloader{},
			},
			want: &PrivateRepo{
				PrivateRepoParams: &PrivateRepoParams{
					S3BinBucket:  "bucket",
					S3BinRegion:  "region",
					S3BinPrefix:  "prefix",
					S3ObjectKey:  &updateMock.ObjectKey{},
					creds:        goodCreds,
					s3svc:        &updateMock.S3API{},
					s3downloader: &updateMock.Downloader{},
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

func TestPrivateRepo_GetAvailableVersions(t *testing.T) {
	timeMustParse := func(val string) *time.Time {
		t, err := time.Parse(time.RFC3339, val)
		if err != nil {
			panic(err)
		}
		return aws.Time(t)
	}

	makeVersions := func(versions ...string) version.Collection {
		col := version.Collection{}
		for _, v := range versions {
			ver, _ := version.NewSemver(v)
			col = append(col, ver)
		}
		return col
	}

	type args struct {
		name string
	}
	tests := []struct {
		name    string
		params  *PrivateRepoParams
		args    args
		want    version.Collection
		wantErr bool
	}{
		{
			name: "should error if unable to list objects",
			params: &PrivateRepoParams{
				s3svc: &updateMock.S3API{
					ListObjectsV2Func: func(in *s3.ListObjectsV2Input) (*s3.ListObjectsV2Output, error) {
						return nil, fmt.Errorf("no way jose")
					},
				},
			},
			wantErr: true,
		},
		{
			name: "should error if unable to parse an s3 object key",
			params: &PrivateRepoParams{
				S3ObjectKey: &updateMock.ObjectKey{
					ParseVersionFunc: func(key, name string) (bool, *version.Version, error) {
						return false, nil, fmt.Errorf("beserk")
					},
				},
				s3svc: &updateMock.S3API{
					ListObjectsV2Func: func(in *s3.ListObjectsV2Input) (*s3.ListObjectsV2Output, error) {
						return &s3.ListObjectsV2Output{
							Contents: []*s3.Object{{Key: aws.String("i'm a cashew")}},
						}, nil
					},
				},
			},
			wantErr: true,
		},
		{
			name: "should return available versions",
			params: &PrivateRepoParams{
				S3ObjectKey: &updateMock.ObjectKey{
					ParseVersionFunc: func(key, name string) (bool, *version.Version, error) {
						var v *version.Version
						var err error
						switch key {
						case "cpd/cpd-v0.1.1-darwin-amd64":
							v, err = version.NewSemver("v0.1.1")
						case "cpd/cpd-v0.1.2-darwin-amd64":
							v, err = version.NewSemver("v0.1.2")
						case "cpd/cpd-v0.1.3-darwin-amd64":
							v, err = version.NewSemver("v0.1.3")
						}
						return true, v, err
					},
				},
				s3svc: &updateMock.S3API{
					ListObjectsV2Func: func(in *s3.ListObjectsV2Input) (*s3.ListObjectsV2Output, error) {
						return &s3.ListObjectsV2Output{
							CommonPrefixes: nil,
							IsTruncated:    aws.Bool(false),
							KeyCount:       aws.Int64(400),
							MaxKeys:        aws.Int64(1000),
							Name:           aws.String("cloud-confluent-bin"),
							Prefix:         aws.String("cpd/"),
							Contents: []*s3.Object{
								{
									ETag:         aws.String("\"d541fd9fc90c385830337448747a21c0-8\""),
									Key:          aws.String("cpd/cpd-v0.1.1-darwin-amd64"),
									LastModified: timeMustParse("2018-07-27T19:14:32Z"),
									Size:         aws.Int64(65154324),
									StorageClass: aws.String("STANDARD"),
								},
								{
									ETag:         aws.String("\"abea850567b589272a4f252bd14a58dc-8\""),
									Key:          aws.String("cpd/cpd-v0.1.2-darwin-amd64"),
									LastModified: timeMustParse("2018-08-02T19:14:32Z"),
									Size:         aws.Int64(65154324),
									StorageClass: aws.String("STANDARD"),
								},
								{
									ETag:         aws.String("\"0524a39b7db0bb5de4bfe015dc5cd78c-8\""),
									Key:          aws.String("cpd/cpd-v0.1.3-darwin-amd64"),
									LastModified: timeMustParse("2018-08-12T19:14:32Z"),
									Size:         aws.Int64(65154324),
									StorageClass: aws.String("STANDARD"),
								},
							},
						}, nil
					},
				},
			},
			want: makeVersions("v0.1.1", "v0.1.2", "v0.1.3"),
		},
		{
			name: "should sort versions",
			params: &PrivateRepoParams{
				S3ObjectKey: &updateMock.ObjectKey{
					ParseVersionFunc: func(key, name string) (bool, *version.Version, error) {
						var v *version.Version
						var err error
						switch key {
						case "cpd/cpd-v0.1.1-darwin-amd64":
							v, err = version.NewSemver("v0.1.1")
						case "cpd/cpd-v0.1.2-darwin-amd64":
							v, err = version.NewSemver("v0.1.2")
						case "cpd/cpd-v0.1.3-darwin-amd64":
							v, err = version.NewSemver("v0.1.3")
						}
						return true, v, err
					},
				},
				s3svc: &updateMock.S3API{
					ListObjectsV2Func: func(in *s3.ListObjectsV2Input) (*s3.ListObjectsV2Output, error) {
						return &s3.ListObjectsV2Output{
							CommonPrefixes: nil,
							IsTruncated:    aws.Bool(false),
							KeyCount:       aws.Int64(400),
							MaxKeys:        aws.Int64(1000),
							Name:           aws.String("cloud-confluent-bin"),
							Prefix:         aws.String("cpd/"),
							Contents: []*s3.Object{
								{
									Key: aws.String("cpd/cpd-v0.1.1-darwin-amd64"),
								},
								{
									Key: aws.String("cpd/cpd-v0.1.3-darwin-amd64"),
								},
								{
									Key: aws.String("cpd/cpd-v0.1.2-darwin-amd64"),
								},
							},
						}, nil
					},
				},
			},
			want: makeVersions("v0.1.1", "v0.1.2", "v0.1.3"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.params.Logger = log.New()
			r := &PrivateRepo{
				PrivateRepoParams: tt.params,
				// Need to inject these so tests pass in different environments (e.g., CI)
				goos:   "darwin",
				goarch: "amd64",
			}
			got, err := r.GetAvailableVersions(tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("PrivateRepo.GetAvailableVersions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PrivateRepo.GetAvailableVersions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPrivateRepo_DownloadVersion(t *testing.T) {
	req := require.New(t)
	type args struct {
		name        string
		version     string
		downloadDir string
	}
	tests := []struct {
		name      string
		params    *PrivateRepoParams
		args      args
		wantPath  string
		wantBytes int64
		wantErr   bool
	}{
		{
			name: "should error if creating file fails",
			params: &PrivateRepoParams{
				fs: &mock.PassThroughFileSystem{
					Mock: &mock.FileSystem{
						CreateFunc: func(name string) (pio.File, error) {
							return nil, errors.New("you no can do that")
						},
					},
					FS: &pio.RealFileSystem{},
				},
			},
			wantErr: true,
		},
		{
			name: "should error if download fails",
			params: &PrivateRepoParams{
				S3BinBucket: "bigbucks",
				S3ObjectKey: &updateMock.ObjectKey{
					URLForFunc: func(name, version string) string {
						return "/some/s3/url"
					},
				},
				s3downloader: &updateMock.Downloader{
					DownloadFunc: func(w io.WriterAt, input *s3.GetObjectInput, options ...func(*s3manager.Downloader)) (int64, error) {
						req.Equal("bigbucks", *input.Bucket)
						req.Equal("/some/s3/url", *input.Key)
						return 0, errors.New("no space here")
					},
				},
				fs: &mock.PassThroughFileSystem{
					Mock: &mock.FileSystem{
						CreateFunc: func(name string) (pio.File, error) {
							return &os.File{}, nil
						},
					},
					FS: &pio.RealFileSystem{},
				},
			},
			wantErr: true,
		},
		{
			name: "should download version",
			params: &PrivateRepoParams{
				S3BinBucket: "bigbucks",
				S3ObjectKey: &updateMock.ObjectKey{
					URLForFunc: func(name, version string) string {
						return "/some/s3/url"
					},
				},
				s3downloader: &updateMock.Downloader{
					DownloadFunc: func(w io.WriterAt, input *s3.GetObjectInput, options ...func(*s3manager.Downloader)) (int64, error) {
						req.Equal("bigbucks", *input.Bucket)
						req.Equal("/some/s3/url", *input.Key)
						return 23, nil
					},
				},
				fs: &mock.PassThroughFileSystem{
					Mock: &mock.FileSystem{
						CreateFunc: func(name string) (pio.File, error) {
							return &os.File{}, nil
						},
					},
					FS: &pio.RealFileSystem{},
				},
			},
			args: args{
				name:        "foofighter",
				version:     "9.8.7", // TODO: shouldn't this need a v prefix?
				downloadDir: "backdoor",
			},
			wantPath:  "backdoor/foofighter-v9.8.7-darwin-amd64",
			wantBytes: 23,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.params.Logger = log.New()
			r := &PrivateRepo{
				PrivateRepoParams: tt.params,
				// Need to inject these so tests pass in different environments (e.g., CI)
				goos:   "darwin",
				goarch: "amd64",
			}
			gotPath, gotBytes, err := r.DownloadVersion(tt.args.name, tt.args.version, tt.args.downloadDir)
			if (err != nil) != tt.wantErr {
				t.Errorf("PrivateRepo.DownloadVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotPath != tt.wantPath {
				t.Errorf("PrivateRepo.DownloadVersion() gotPath = %v, wantPath %v", gotPath, tt.wantPath)
			}
			if gotBytes != tt.wantBytes {
				t.Errorf("PrivateRepo.DownloadVersion() gotBytes = %v, wantPath %v", gotBytes, tt.wantBytes)
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

func makeCreds(profile string, val credentials.Value, err error, expired bool) *mockCredsFactory {
	return &mockCredsFactory{allCreds: []credsAssert{
		{expectProfile: profile, provider: &mockCredentialsProvider{val, err, expired}},
	}}
}
