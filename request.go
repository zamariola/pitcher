package pitcher

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/google/uuid"
)

const (
	hostKey = "host"
	jwtKey  = "jwt_token"
)

const (
	authorizationKey       = "Authorization"
	contentApplicationJson = "application/json"
	contentType            = "Content-Type"
)

var verbose int

type PreProcessorFunc func(*Request, Session) error

type AssertionFunc func(*Response) bool

type PostProcessorFunc func(*Request, *Response, Session) error

type StringParserFunc func(string) string

type Step struct {
	PreProcs   []PreProcessorFunc
	Request    *Request
	Assertions []AssertionFunc
	PostProcs  []PostProcessorFunc
}

func (s Step) WithPreProcessors(preProcessors ...PreProcessorFunc) Step {
	return Step{
		Request:    s.Request,
		Assertions: s.Assertions,
		PostProcs:  s.PostProcs,
		PreProcs:   preProcessors,
	}
}

func (s Step) WithPostProcessors(postProcessors ...PostProcessorFunc) Step {
	return Step{
		Request:    s.Request,
		Assertions: s.Assertions,
		PostProcs:  postProcessors,
		PreProcs:   s.PreProcs,
	}
}

type Request struct {
	Method      string
	Host        string
	Path        string
	Body        string
	Query       map[string]string
	ContentType string
	Headers     http.Header
}

func Success(request *Request) Step {
	return Step{
		Request: request,
		Assertions: []AssertionFunc{
			SuccessAssertion,
		},
	}
}

func GET(path string) Step {
	return Success(&Request{
		Method: "GET",
		Path:   path,
		Query:  map[string]string{},
	})
}

func POST(path string, body string, contentType string) Step {
	return Success(&Request{
		Method:      "POST",
		Path:        path,
		Body:        body,
		ContentType: contentType,
		Query:       map[string]string{},
	})
}

func PUT(path string, body string, contentType string) Step {
	return Success(
		&Request{
			Method:      "PUT",
			Path:        path,
			Body:        body,
			ContentType: contentType,
			Query:       map[string]string{},
		})
}

func PATCH(path string, body string, contentType string) Step {
	return Success(
		&Request{
			Method:      "PATCH",
			Path:        path,
			Body:        body,
			ContentType: contentType,
			Query:       map[string]string{},
		})
}

func DELETE(path string, body string, contentType string) Step {
	return Success(&Request{
		Method: "DELETE",
		Path:   path,
		Query:  map[string]string{},
	})
}

type Response struct {
	StatusCode int
	Body       string
	Headers    http.Header
}

type Client struct {
	client          *http.Client
	session         Session
	globalPreProcs  []PreProcessorFunc
	globalPostProcs []PostProcessorFunc
	namedSteps      map[string]Step
}

func NewClient() *Client {
	s := NewMemoryRWSession(make(map[string]string))

	return &Client{
		client:     http.DefaultClient,
		session:    s,
		namedSteps: map[string]Step{},
	}
}

func NewClientWithSession(session Session) *Client {
	return NewClientWithProcessors(
		session,
		[]PreProcessorFunc{},
		[]PostProcessorFunc{},
	)
}

func NewClientWithProcessors(
	session Session,
	preProcs []PreProcessorFunc,
	postProcs []PostProcessorFunc,
) *Client {
	return NewCustomClient(
		http.DefaultClient,
		session,
		preProcs,
		postProcs,
	)
}

func NewCustomClient(
	client *http.Client,
	session Session,
	preProcs []PreProcessorFunc,
	postProcs []PostProcessorFunc,
) *Client {
	return &Client{
		client:          client,
		session:         session,
		globalPreProcs:  preProcs,
		globalPostProcs: postProcs,
		namedSteps:      map[string]Step{},
	}
}

func (c *Client) SetTransport(t *http.Transport) *Client {
	c.client.Transport = t
	return c
}

func (c *Client) Add(id string, step Step) {
	c.namedSteps[id] = step
}

func (c *Client) Parse() {

	sessionValues := make(KeyValuePairs)
	flag.Var(sessionValues, "s", "Specify session key=value pairs")
	flag.IntVar(&verbose, "v", 1, "set verbose log level (0=response only; 1=info; 2+=debug)")
	flag.Parse()

	//setting log level

	var level = slog.LevelInfo

	if verbose == 0 {
		level = slog.LevelError
	} else if verbose >= 2 {
		level = slog.LevelDebug
	}
	opts := &slog.HandlerOptions{
		Level: level,
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, opts))
	slog.SetDefault(logger)
	cmdSteps := flag.Args()

	if len(cmdSteps) == 0 {
		keys := make([]string, 0, len(c.namedSteps))

		for k := range c.namedSteps {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		slog.Info("No step(s) provided. Please provide at least on step:")

		for _, v := range keys {
			slog.Info(v)
		}
		return
	}

	// add cmd session values
	for k, v := range sessionValues {
		c.session.Put(k, v)
	}

	// do the sequential call
	steps := make([]Step, 0)

	for _, v := range cmdSteps {
		s, ok := c.namedSteps[v]

		if !ok {
			slog.Error("Unable to find provided step", "step", v)
			os.Exit(1)
		}
		steps = append(steps, s)
	}

	slog.Debug("Executing request", "steps", cmdSteps, "session", c.session)
	c.Do(steps...)
}

func (c *Client) Do(steps ...Step) ([]*Response, error) {

	responses := []*Response{}

	for _, step := range steps {
		r, err := c.runStep(step)

		if err != nil {
			return responses, err
		}
		responses = append(responses, r)
	}
	return responses, nil
}

func (c *Client) Test(t *testing.T, steps ...Step) []*Response {

	r, err := c.Do(steps...)

	if err != nil {
		t.Error(err)
	}
	return r
}

func (c *Client) runStep(step Step) (*Response, error) {

	//Prepare the request
	req := step.Request

	if req.Headers == nil {
		req.Headers = http.Header{}
	}

	//Global PreProcessors
	for _, proc := range c.globalPreProcs {
		proc(req, c.session)
	}

	//PreProcessors
	for _, proc := range step.PreProcs {
		proc(req, c.session)
	}

	//String parsers

	var parsers = []StringParserFunc{
		parseUUID,
		c.parseSessionKeys,
	}

	for _, parser := range parsers {
		req.Body = parser(req.Body)
		req.Host = parser(req.Host)
		req.Path = parser(req.Path)

		for k, v := range req.Query {
			req.Query[k] = parser(v)
		}

	}

	//Fallback null/invalid host to session variables
	host := req.Host

	if _, err := url.Parse(host); len(host) == 0 || err != nil {
		host, _ = c.session.Get(hostKey)
	}

	urlP, err := url.JoinPath(host, req.Path)

	if err != nil {
		return &Response{}, err
	}

	//Do the request
	var reqBody io.Reader

	if len(req.Body) > 0 {
		reqBody = strings.NewReader(req.Body)

		if len(req.ContentType) == 0 {
			req.ContentType = contentApplicationJson
		}
	}

	request, err := http.NewRequest(req.Method, urlP, reqBody)
	request.Header = req.Headers

	//Add Query params
	if len(req.Query) > 0 {
		params := url.Values{}

		for key, val := range req.Query {
			params.Add(key, val)
		}
		request.URL.RawQuery = params.Encode()
	}

	if len(req.ContentType) > 0 {
		request.Header.Add(contentType, contentApplicationJson)
	}

	if err != nil {
		return &Response{}, err
	}

	resp, err := c.client.Do(request)
	if err != nil {
		return &Response{}, err
	}

	respBody, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return &Response{}, err
	}

	response := &Response{
		StatusCode: resp.StatusCode,
		Body:       string(respBody),
		Headers:    resp.Header,
	}

	//Global Post Processors
	for _, proc := range c.globalPostProcs {
		if err := proc(req, response, c.session); err != nil {
			return response, err
		}
	}

	//Post processors
	for _, proc := range step.PostProcs {
		if err := proc(req, response, c.session); err != nil {
			return response, err
		}
	}

	//Assertions
	for _, proc := range step.Assertions {
		if valid := proc(response); !valid {
			return response, errors.New("invalid response for assertion")
		}
	}

	return response, nil
}

func (c *Client) parseSessionKeys(body string) string {

	re := regexp.MustCompile(`\$\{(.*?)\}`)

	placeholders := re.FindAllStringSubmatch(body, -1)

	for _, placeholder := range placeholders {
		if len(placeholder) > 1 {
			replacer, key := placeholder[0], placeholder[1]

			if value, ok := c.session.Get(key); ok {
				body = strings.Replace(body, replacer, value, 1)
			}
		}
	}

	return body
}

type KeyValuePairs map[string]string

func (kv KeyValuePairs) String() string {
	var pairs []string
	for k, v := range kv {
		pairs = append(pairs, fmt.Sprintf("%s=%s", k, v))
	}
	return strings.Join(pairs, ", ")
}

func (kv KeyValuePairs) Set(value string) error {
	parts := strings.SplitN(value, "=", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid format, expected key=value")
	}
	kv[parts[0]] = parts[1]
	return nil
}

func parseUUID(body string) string {
	return strings.Replace(body, "${randomUUID}", uuid.New().String(), -1)
}
