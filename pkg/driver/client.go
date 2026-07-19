package driver

import (
	"net/http"

	"github.com/go-resty/resty/v2"
)

// sentinelEmptyUA is a non-empty marker that signals the User-Agent header
// should be stripped before the HTTP request is sent. It prevents resty v2's
// middleware from overriding an empty UA with its default value.
const sentinelEmptyUA = "\x00__EMPTY_UA__"

// Pan115Client driver client
type Pan115Client struct {
	Client            *resty.Client
	Request           *resty.Request
	UserID            int64
	Userkey           string
	UploadMetaInfo    *UploadMetaInfo
	UseInternalUpload bool
}

// New creates Client with customized options.
func New(opts ...Option) *Pan115Client {
	c := &Pan115Client{
		Client: resty.New(),
	}

	// Hook 1: Before resty's middleware — detect empty UA and replace with sentinel.
	// This runs before parseRequestHeader, so we catch empty UA at the request level.
	// For client-level empty UA (set via SetHeader on c.Client), we check client.Header
	// because client headers haven't been merged into the request yet at this point.
	c.Client.OnBeforeRequest(func(client *resty.Client, r *resty.Request) error {
		if vals, exists := r.Header["User-Agent"]; exists && len(vals) > 0 && vals[0] == "" {
			r.Header.Set("User-Agent", sentinelEmptyUA)
			return nil
		}
		if vals, exists := client.Header["User-Agent"]; exists && len(vals) > 0 && vals[0] == "" {
			client.SetHeader("User-Agent", sentinelEmptyUA)
		}
		return nil
	})

	// Hook 2: After resty's middleware — strip sentinel before HTTP send.
	c.Client.SetPreRequestHook(func(client *resty.Client, req *http.Request) error {
		if val := client.Header.Get("User-Agent"); val == sentinelEmptyUA {
			client.SetHeader("User-Agent", "")
		}
		if req.Header.Get("User-Agent") == sentinelEmptyUA {
			req.Header.Set("User-Agent", "")
		}
		return nil
	})

	if len(opts) > 0 {
		for _, optFunc := range opts {
			optFunc(c)
		}
	}
	return c
}

// Default creates an Client with default settings.
func Default() *Pan115Client {
	return New(UA())
}

// Defalut is deprecated: use Default instead. This function exists for backward compatibility.
func Defalut() *Pan115Client {
	return Default()
}

func (c *Pan115Client) SetHttpClient(httpClient *http.Client) *Pan115Client {
	c.Client = resty.NewWithClient(httpClient)
	return c
}

func (c *Pan115Client) SetUserAgent(userAgent string) *Pan115Client {
	c.Client.SetHeader("User-Agent", userAgent)
	return c
}

func (c *Pan115Client) SetCookies(cs ...*http.Cookie) *Pan115Client {
	c.Client.SetCookies(cs)
	return c
}

func (c *Pan115Client) SetDebug(d bool) *Pan115Client {
	c.Client.SetDebug(d)
	return c
}

func (c *Pan115Client) EnableTrace() *Pan115Client {
	c.Client.EnableTrace()
	return c
}

func (c *Pan115Client) SetProxy(proxy string) *Pan115Client {
	c.Client.SetProxy(proxy)
	return c
}

func (c *Pan115Client) NewRequest() *resty.Request {
	c.Request = c.Client.R()
	return c.Request
}

func (c *Pan115Client) GetRequest() *resty.Request {
	if c.Request != nil {
		return c.Request
	}
	return c.NewRequest()
}
