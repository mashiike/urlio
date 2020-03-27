package urlio

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"

	"cloud.google.com/go/storage"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"google.golang.org/api/option"
)

// Constructor provides the construction function
type Constructor interface {
	NewReader(*url.URL) (io.ReadCloser, error)
}

// ConstructorMap represents the correspondence between Scheme and Constructor
type ConstructorMap map[string]Constructor

var (
	std     ConstructorMap
	builtin ConstructorMap

	stdS3   *S3
	stdGS   *GS
	stdFile *File
	stdHTTP *HTTP
)

func init() {
	stdS3 = NewS3()
	stdGS = NewGS()
	stdFile = NewFile()
	stdHTTP = NewHTTP()
	builtin = ConstructorMap{
		"s3":    stdS3,
		"gs":    stdGS,
		"file":  stdFile,
		"":      stdFile,
		"http":  stdHTTP,
		"https": stdHTTP,
	}
	std = New()
}

func New() ConstructorMap {
	m := ConstructorMap{}
	m.Constractors(builtin)
	return m
}

// Constractors adds the elements of the argument map to the function map of the constructor.
// Must be called before building a io.ReadCloser
func (m ConstructorMap) Constractors(constractorMap ConstructorMap) {
	for scheme, constractor := range constractorMap {
		m[scheme] = constractor
	}
}

// NewReader returns a new io.ReadCloser according to the URL scheme of the source.
func (m ConstructorMap) NewReader(src *url.URL) (io.ReadCloser, error) {
	if constructor, ok := m[src.Scheme]; ok {
		return constructor.NewReader(src)
	}
	return nil, fmt.Errorf("source URL scheme not supported: %s", src.String())
}

// S3 provides AWS S3 constractor func
type S3 struct {
	mu   sync.Mutex
	svc  *s3.S3
	conf *aws.Config
}

func NewS3() *S3 {
	return &S3{
		conf: aws.NewConfig().WithRegion(os.Getenv("AWS_REGION")),
	}
}

func (c *S3) getSvc() (*s3.S3, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.svc != nil {
		return c.svc, nil
	}
	sess, err := session.NewSessionWithOptions(session.Options{
		Config:            *c.conf,
		SharedConfigState: session.SharedConfigEnable,
	})
	if err != nil {
		return nil, err
	}
	c.svc = s3.New(sess)
	return c.svc, nil
}

// Config sets AWS session Configure
func (c *S3) Config(cfgs ...*aws.Config) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.conf.MergeIn(cfgs...)
	c.svc = nil
}

// NewReader returns a new io.ReadCloser according to s3 resource.
func (c *S3) NewReader(src *url.URL) (io.ReadCloser, error) {
	svc, err := c.getSvc()
	if err != nil {
		return nil, err
	}
	key := strings.TrimLeft(src.Path, "/")
	result, err := svc.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(src.Host),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	return result.Body, nil
}

// GS provides GCP CloucStorage constractor func
type GS struct {
	opts []option.ClientOption
}

func NewGS(opts ...option.ClientOption) *GS {
	return &GS{
		opts: opts,
	}
}

// Config sets GCP client Option
func (c *GS) Config(opts ...option.ClientOption) {
	c.opts = append(c.opts, opts...)
}

// NewReader returns a new io.ReadCloser according to gs resource.
func (c *GS) NewReader(src *url.URL) (io.ReadCloser, error) {
	var err error
	trimPath := strings.Trim(src.Path, "/")

	ctx := context.Background()
	client, err := storage.NewClient(ctx, c.opts...)
	if err != nil {
		return nil, err
	}
	return client.Bucket(src.Host).Object(trimPath).NewReader(ctx)
}

type File struct{}

func NewFile() *File {
	return &File{}
}

// NewReader returns a new io.ReadCloser according to local FileSystem resource.
func (c *File) NewReader(src *url.URL) (io.ReadCloser, error) {
	if src.Opaque != "" {
		return os.Open(src.Opaque)
	}
	return os.Open(src.Path)
}

// HTTP provides HTTP/HTTPS request resource constractor func
type HTTP struct {
	agentName   string
	checkStatus bool
	client      *http.Client
}

func NewHTTP() *HTTP {
	return &HTTP{
		agentName:   "urlio",
		client:      http.DefaultClient,
		checkStatus: false,
	}
}

type HTTPOption interface {
	Apply(*HTTP)
}

// Config sets HTTPOptions
func (c *HTTP) Config(opts ...HTTPOption) {
	for _, opt := range opts {
		opt.Apply(c)
	}
}

type withHTTPClient http.Client

func (o *withHTTPClient) Apply(c *HTTP) {
	c.client = (*http.Client)(o)
}

// WithHTTPClient is a HTTP/HTTPS scheme option. set *http.Client
func WithHTTPClient(client *http.Client) HTTPOption {
	return (*withHTTPClient)(client)
}

type withUserAgent string

func (o withUserAgent) Apply(c *HTTP) {
	c.agentName = string(o)
}

// WithUserAgent is a HTTP/HTTPS scheme option. set user-agent
func WithUserAgent(agentName string) HTTPOption {
	return withUserAgent(agentName)
}

type withCheckStatus bool

func (o withCheckStatus) Apply(c *HTTP) {
	c.checkStatus = bool(o)
}

// WithCheckStatus is HTTP/HTTPS scheme option. if set true, status 4xx, 5xx,... not 200 is error
func WithCheckStatus(check bool) HTTPOption {
	return withCheckStatus(check)
}

// NewReader returns a new io.ReadCloser according to http/https resource.
func (c *HTTP) NewReader(src *url.URL) (io.ReadCloser, error) {
	req, err := http.NewRequest("GET", src.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("User-Agent", c.agentName)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	if c.checkStatus && resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("request %s, response %s", src.String(), resp.Status)
	}
	return resp.Body, nil

}

// Constractors adds the elements of the argument map to the function map of the constructor.
// Must be called before building a io.ReadCloser
func Constructors(constractorMap ConstructorMap) {
	std.Constractors(constractorMap)
}

// NewReader returns a new io.ReadCloser according to the URL scheme of the source.
func NewReader(src *url.URL) (io.ReadCloser, error) {
	return std.NewReader(src)
}

// S3Config sets AWS session Configure
func S3Config(cfgs ...*aws.Config) {
	stdS3.Config(cfgs...)
}

// GSConfig sets GCP client Option
func GSConfig(opts ...option.ClientOption) {
	stdGS.Config(opts...)
}

// HTTPConfig sets HTTPOptions
func HTTPConfig(opts ...HTTPOption) {
	stdHTTP.Config(opts...)
}

//  MustParse as url.Parse. if error occted panic.
func MustParse(rawurl string) *url.URL {
	u, err := url.Parse(rawurl)
	if err != nil {
		panic(err)
	}
	return u

}
