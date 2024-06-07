package pitcher

import (
	"log/slog"

	"github.com/tidwall/gjson"
)

func LogStepProcessor(req *Request, resp *Response, session Session) error {
	slog.Info("Executing", "method", req.Method, "path", req.Path, "statusCode", resp.StatusCode)
	return nil
}

func LogPayloadProcessor(req *Request, resp *Response, session Session) error {
	slog.Info("Executing", "method", req.Method, "path", req.Path, "statusCode", resp.StatusCode, "body", resp.Body)
	return nil
}

func Extract(key, path string) PostProcessorFunc {
	return func(req *Request, resp *Response, session Session) error {

		value := gjson.Get(resp.Body, path)

		if !value.Exists() {
			slog.Info("unable to find value in the json response", "key", key)
			return nil
		}

		session.Put(key, value.String())
		return nil
	}
}
