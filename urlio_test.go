package urlio_test

import (
	"io/ioutil"
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
	initHTTP()
	m.Run()
}

func TestNewReader(t *testing.T) {
	testdataAbsPath, _ := filepath.Abs("./testdata")
	//success
	cases := [][]string{
		{"s3://bucket.example.com/object.txt", "hoge\n"},
		{"gs://bucket/object.txt", "fuga\n"},
		{"file:./testdata/file/data.txt", "piyo\n"},
		{"file://" + testdataAbsPath + "/file/data.txt", "piyo\n"},
		{"file://./testdata/file/data.txt", "piyo\n"},
		{"./testdata/file/data.txt", "piyo\n"},
		{testdataAbsPath + "/file/data.txt", "piyo\n"},
		{"http://web.example.com/data.txt", "tora\n"},
		{"https://web.example.com/data.txt", "tora\n"},
	}
	for _, c := range cases {
		t.Run(c[0], func(t *testing.T) {
			reader, err := urlio.NewReader(urlio.MustParse(c[0]))
			if err != nil {
				t.Errorf("can not create reader: %s", err)
				return
			}
			defer reader.Close()
			b, err := ioutil.ReadAll(reader)
			if err != nil {
				t.Errorf("can not read data: %s", err)
				return
			}
			if string(b) != c[1] {
				t.Errorf("unexpected data. got = %q, expected = %q", string(b), c[1])
			}
		})
	}

	//failed
	fcases := []string{
		"s3://bucket.example.com/not_found.txt",
		"gs://bucket/not_found.txt",
		"file:./testdata/file/not_found.txt",
		"file://" + testdataAbsPath + "/file/not_found.txt",
		"http://web.example.com/not_found.txt",
		"https://web.example.com/not_found.txt",
		"local:///invalid_scheme.txt",
	}
	for _, fc := range fcases {
		t.Run(fc, func(t *testing.T) {
			reader, err := urlio.NewReader(urlio.MustParse(fc))
			if err == nil {
				reader.Close()
				t.Errorf("NewReader must failed")
			}
		})
	}
}

func initS3() {
	base, _ := filepath.Abs("testdata/s3")
	urlio.S3Config(
		&aws.Config{
			HTTPClient:       internal.NewProxyClient(base, false),
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
		option.WithHTTPClient(internal.NewProxyClient(base, false)),
	)
}

func initHTTP() {
	base, _ := filepath.Abs("testdata/http")
	urlio.HTTPConfig(
		urlio.WithHTTPClient(internal.NewProxyClient(base, true)),
		urlio.WithUserAgent("urlio/test"),
		urlio.WithCheckStatus(true),
	)
}
