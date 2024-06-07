package pitcher

import "net/http"

func SuccessAssertion(resp *Response) bool {
	return resp.StatusCode >= 200 && resp.StatusCode <= 299
}

func NotFoundAssertion(resp *Response) bool {
	return resp.StatusCode == http.StatusNotFound
}
