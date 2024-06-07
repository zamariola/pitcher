package pitcher

import (
	"errors"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"

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

func NewStep(request *Request) Step {
	return Step{
		Request: request,
	}
}

type Request struct {
	Method      string
	Host        string
	Path        string
	Body        string
	ContentType string
	Headers     http.Header
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
}

func NewClient() *Client {
	s := NewMemoryRWSession(make(map[string]string))

	return &Client{
		client:  http.DefaultClient,
		session: s,
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
	}
}

func (c *Client) Do(steps ...Step) error {

	for _, step := range steps {
		if err := c.runStep(step); err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) runStep(step Step) error {

	//Prepare the request
	req := step.Request

	var parsers = []StringParserFunc{
		parseUUID,
		c.parseSessionKeys,
	}

	for _, parser := range parsers {
		req.Body = parser(req.Body)
		req.Host = parser(req.Host)
		req.Path = parser(req.Path)
	}

	if req.Headers == nil {
		req.Headers = http.Header{}
	}

	//PreProcessors
	for _, proc := range step.PreProcs {
		proc(req, c.session)
	}

	//Fallback null/invalid host to session variables
	host := req.Host

	if _, err := url.Parse(host); len(host) == 0 || err != nil {
		host, _ = c.session.Get(hostKey)
	}

	urlP, err := url.JoinPath(host, req.Path)

	if err != nil {
		return err
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

	if len(req.ContentType) > 0 {
		request.Header.Add(contentType, contentApplicationJson)
	}

	if err != nil {
		return err
	}

	resp, err := c.client.Do(request)
	if err != nil {
		return err
	}

	respBody, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return err
	}

	response := &Response{
		StatusCode: resp.StatusCode,
		Body:       string(respBody),
		Headers:    resp.Header,
	}

	//Global Post Processors
	for _, proc := range c.globalPostProcs {
		if err := proc(req, response, c.session); err != nil {
			return err
		}
	}

	//Post processors
	for _, proc := range step.PostProcs {
		if err := proc(req, response, c.session); err != nil {
			return err
		}
	}

	//Assertions
	for _, proc := range step.Assertions {
		if valid := proc(response); !valid {
			return errors.New("invalid response for assertion")
		}
	}

	return nil
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

func parseUUID(body string) string {
	return strings.Replace(body, "${randomUUID}", uuid.New().String(), -1)
}
