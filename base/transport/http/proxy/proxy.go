package proxy

import (
	"context"
	"net"
	net_http "net/http"
	"net/url"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/apoorvprecisely/go-base/base/endpoint"
	"github.com/apoorvprecisely/go-base/base/log"
)

const defaultUserAgent = "Mozart-[go-dialer]"

var (
	hopHeaders = []string{
		"Connection",
		"Proxy-Connection",
		"Keep-Alive",
		"Proxy-Authenticate",
		"Proxy-Authorization",
		"Te",
		"Trailer",
		"Transfer-Encoding",
		"Upgrade",
	}
)

type (
	// RequestOption defines means to modify request
	RequestOption func(req *net_http.Request) error

	// ResponseOption defines means to modify response
	ResponseOption func(res *net_http.Response) error

	// Proxy defines a reverse proxy for wingman
	Proxy struct {
		reqopts []RequestOption
		resopts []ResponseOption

		logger log.Logger

		dialer net_http.RoundTripper

		path string
	}

	// ProxyOption is set of options which can modify proxy
	ProxyOption func(*Proxy)
)

func deleteHopHeaderRequestOption(req *net_http.Request) (err error) {
	for _, h := range hopHeaders {
		req.Header.Del(h)
	}

	return err
}

func deleteHopHeaderResponseOption(res *net_http.Response) (err error) {
	for _, h := range hopHeaders {
		res.Header.Del(h)
	}

	return err
}
func newdirectorRequestOption(uri *url.URL) func(req *net_http.Request) error {
	return func(req *net_http.Request) error {
		if uri == nil {
			return errors.New("uri is nil")
		}

		query := buildQuery(
			req.URL.Query().Encode(),
			uri.Query().Encode(),
		)

		// uri
		req.URL.Scheme = uri.Scheme
		req.URL.Host = uri.Host
		req.URL.RawQuery = query

		// host
		req.Host = uri.Host

		// header
		req.Header.Set("Host", uri.Host)
		return nil
	}
}

func newUserAgentRequestOption(userAgent string) func(req *net_http.Request) error {
	return func(req *net_http.Request) error {
		if userAgent != *new(string) &&
			req.Header.Get("User-Agent") == *new(string) {
			req.Header.Set("User-Agent", userAgent)
		}
		return nil
	}
}

func deleteConnectionResponseOption(res *net_http.Response) (err error) {
	cn := res.Header.Get("Connection")
	if cn != *new(string) {
		for _, f := range strings.Split(cn, ",") {
			if f = strings.TrimSpace(f); f != *new(string) {
				res.Header.Del(f)
			}
		}
	}
	return err
}

func newRequest(cx context.Context, req *net_http.Request) *net_http.Request {
	nr := req.WithContext(cx)

	// body
	if req.ContentLength == 0 {
		nr.Body = nil
	}

	// header
	nr.Header = make(net_http.Header, len(req.Header))
	for k, vs := range req.Header {
		nvs := make([]string, len(vs))
		copy(nvs, vs)

		nr.Header[k] = nvs
	}

	// cleanup
	nr.Close = false

	// x-forwarded-for
	ip, _, err := net.SplitHostPort(req.RemoteAddr)
	if err == nil {
		prev, ok := nr.Header["X-Forwarded-For"]
		if ok {
			ip = strings.Join(prev, ",") + "," + ip
		}

		nr.Header.Set("X-Forwarded-For", ip)
	}
	return nr
}

func buildPath(fromConfig string, req *net_http.Request) (string, error) {
	prefix := fromConfig
	suffix := req.URL.Path

	return buildPathVals(prefix, suffix), nil
}

func buildPathVals(prefix, suffix string) string {
	hp := strings.HasSuffix(prefix, "/")
	hs := strings.HasPrefix(suffix, "/")

	switch {
	case hp && hs:
		return prefix + suffix[1:]
	case !hp && !hs:
		return prefix + "/" + suffix
	default:
		return prefix + suffix
	}
}

func buildQuery(fromRequest, fromConfig string) string {
	if fromRequest == *new(string) && fromConfig == *new(string) {
		return fromConfig + fromRequest
	}

	return fromConfig + "&" + fromRequest
}

func requestOptions(req *net_http.Request, options ...RequestOption) error {
	for _, opt := range options {
		err := opt(req)
		if err != nil {
			return err
		}
	}
	return nil
}

func responseOptions(res *net_http.Response, options ...ResponseOption) error {
	for _, opt := range options {
		err := opt(res)
		if err != nil {
			return err
		}
	}
	return nil
}

// HandlerFunc returns endpoint for reverse proxy
func (pr *Proxy) HandlerFunc() func(context.Context, *net_http.Request) (*net_http.Response, error) {
	return func(
		cx context.Context,
		req *net_http.Request,
	) (*net_http.Response, error) {
		var (
			outreq *net_http.Request
			outres *net_http.Response

			path string
			err  error
		)

		// context in request shouldn't use the `cx context.Context`
		// supplied by the method above. That context is the application
		// context and only contains details pertaining the application
		// flow.
		// context carrier for request is request.Context()
		outreq = newRequest(
			req.Context(),
			req,
		)

		path, err = buildPath(pr.path, req)
		if err != nil {
			return nil, errors.Wrap(err, "build path failed")
		}

		err = requestOptions(outreq, append(
			pr.reqopts, func(req *net_http.Request) error {
				req.URL.Path = path
				return nil
			})...)
		if err != nil {
			return nil, errors.Wrap(
				err, "request options failed",
			)
		}

		pr.logger.Debug("Dialing",
			log.String("Host", outreq.URL.Host),
			log.String("Path", outreq.URL.Path),
			log.String("Query", outreq.URL.RawQuery),
			log.String("RequestID", outreq.Header.Get("x-request-id")),
		)

		outres, err = pr.dialer.RoundTrip(outreq)
		if err != nil {
			return nil, errors.Wrap(
				err, "dial request to downstream failed",
			)
		}

		pr.logger.Debug("Dialed Host",
			log.String("Host", outreq.URL.Host),
			log.String("RequestID", outreq.Header.Get("x-request-id")),
			log.Error(err),
			log.String("Status", outres.Status),
			log.Int("StatusCode", outres.StatusCode),
		)

		err = responseOptions(outres, pr.resopts...)
		if err != nil {
			return nil, errors.Wrap(
				err, "response options failed",
			)
		}

		return outres, nil
	}
}

// ProxyWithCustomTransport provides option to set custom roundtripper for the
// reverse proxy
func ProxyWithCustomTransport(rt net_http.RoundTripper) ProxyOption {
	return func(pr *Proxy) {
		pr.dialer = rt
	}
}

// ProxyWithRequestOptions provies option to append custom RequestOption for
// the reverse proxy
func ProxyWithRequestOptions(fns ...RequestOption) ProxyOption {
	return func(pr *Proxy) {
		pr.reqopts = append(pr.reqopts, fns...)
	}
}

// ProxyWithResponseOptions provies option to append custom ResponseOptions for
// the reverse proxy
func ProxyWithResponseOptions(fns ...ResponseOption) ProxyOption {
	return func(pr *Proxy) {
		pr.resopts = append(pr.resopts, fns...)
	}
}

// ProxyWithModifiedTransport provides option to customize the transport used
// in dialing downstream
func ProxyWithModifiedTransport(
	dialerTimeout time.Duration,
	dialerKeepAlive time.Duration,
	idleConnTimeout time.Duration,
	maxIdle int,
) ProxyOption {
	return func(pr *Proxy) {
		pr.dialer = &net_http.Transport{
			Proxy: net_http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   dialerTimeout,
				KeepAlive: dialerKeepAlive,
				DualStack: true,
			}).DialContext,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          maxIdle,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		}
	}
}

// NewProxy returns a new reverse proxy
func NewProxy(
	logger log.Logger,
	downstream string,
	options ...ProxyOption,
) (*Proxy, error) {

	uri, err := url.Parse(downstream)
	if err != nil {
		return nil, errors.Wrap(err, "parse url failed")
	}

	logger.Debug("URL",
		log.String("downstream", downstream),
		log.String("scheme", uri.Scheme),
		log.String("host", uri.Host),
		log.Reflect("uri", uri),
	)

	proxy := &Proxy{
		reqopts: []RequestOption{
			deleteHopHeaderRequestOption,
			newdirectorRequestOption(uri),
			newUserAgentRequestOption(defaultUserAgent),
		},

		resopts: []ResponseOption{
			deleteConnectionResponseOption,
			deleteHopHeaderResponseOption,
		},

		logger: logger,
		path:   uri.Path,
		dialer: net_http.DefaultTransport,
	}

	for _, opt := range options {
		opt(proxy)
	}

	return proxy, nil
}

// NewProxyEndpoint returns an Endpoint which handles an incoming http.Request
func NewProxyEndpoint(
	logger log.Logger,
	downstream string,
	options ...ProxyOption,
) (endpoint.Endpoint, error) {
	prx, err := NewProxy(logger, downstream, options...)
	if err != nil {
		return nil, errors.Wrap(err, "create proxy object failed")
	}

	return endpoint.Endpoint(func(
		cx context.Context, req interface{},
	) (res interface{}, err error) {
		rq, ok := req.(*net_http.Request)
		if !ok {
			return nil, errors.New("'req' should be net/http.Request")
		}

		return prx.HandlerFunc()(cx, rq)
	}), nil
}
