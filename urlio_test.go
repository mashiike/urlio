package urlio_test

import (
	"path/filepath"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/mashiike/urlio"
	"github.com/mashiike/urlio/internal"
	"google.golang.org/api/option"
)

func TestMain(m *testing.M) {
	initS3()
	initGS()
	m.Run()
}

func TestNewReader(t *testing.T) {
	urls := []string{
		"s3://bucket.example.com/object.txt",
		"gs://bucket/object.txt",
	}
	for _, u := range urls {
		t.Run(u, func(t *testing.T) {
			reader, err := urlio.NewReader(urlio.MustParse(u))
			if err != nil {
				t.Fatal(err)
			}
			defer reader.Close()
		})
	}
}

func initS3() {
	base, _ := filepath.Abs("testdata/s3")
	urlio.S3Config(
		&aws.Config{
			HTTPClient:       internal.NewProxyClient(base),
			DisableSSL:       aws.Bool(true),
			S3ForcePathStyle: aws.Bool(true),
			Region:           aws.String("ap-northeast-1"),
		},
	)
}

func initGS() {
	base, _ := filepath.Abs("testdata/gs")
	urlio.GSConfig(
		option.WithoutAuthentication(),
		option.WithHTTPClient(internal.NewProxyClient(base)),
	)
}
