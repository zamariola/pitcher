package pitcher

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/tidwall/gjson"
)

func LogStepProcessor(req *Request, resp *Response, session Session) error {
	slog.Info("Executing", "method", req.Method, "path", req.Path, "statusCode", resp.StatusCode)
	return nil
}

func LogPayloadProcessor(req *Request, resp *Response, session Session) error {

	if strings.Contains(resp.Headers.Get("Content-Type"), "application/json") {
		var rawJSON interface{}
		json.Unmarshal([]byte(resp.Body), &rawJSON)

		formattedJSON, _ := json.MarshalIndent(rawJSON, "", "  ")

		fmt.Println(string(formattedJSON))
		return nil

	}

	fmt.Println(resp.Body)
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
