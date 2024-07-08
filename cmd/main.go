package main

import (
	"flag"
	"fmt"

	pitcher "github.com/zamariola/pitcher"
)

var local = map[string]string{
	"host": "https://jsonplaceholder.typicode.com",
}

var dev = map[string]string{
	"host": "https://SOME_DEV_URL.typicode.com",
}

func main() {

	// Get the map of parameters for the desired environment
	envParameters := parametersFromEnv()

	// Define the client with the global processors
	client := pitcher.NewClientWithProcessors(

		//Read and write session that expires once it finishes
		pitcher.NewMemoryRWSession(envParameters),

		//Slice of global (every request) Pre Processors that set ups the session and variables
		//if neeeded
		[]pitcher.PreProcessorFunc{

			//Pre processor that reads the "jwt_token" variable and adds to the Request Header
			pitcher.JWTAuth,
		},

		//Slice of Post Processor that manipulates the response and could extract values from
		//response to the session for reusage
		[]pitcher.PostProcessorFunc{

			//Post processor to log the step result
			pitcher.LogStepProcessor,
		},
	)

	_, err := client.Do(
		//Fluent API to create a request
		//pitcher.GET("<ur>").WithPreProcessors(...PreProcessors).WithPostProcessors(...PostProcessors)
		pitcher.GET("/posts").
			// Pre processors that prepares the session or update the request before the execution
			WithPreProcessors(
				pitcher.UpdateSession("jwtToken", "eyJhbGciOiJIUzI1NiIsInR5cC..."),
				pitcher.JWTAuth,
			).
			WithPostProcessors(
				//Post processor that extracts the json value from the path 0.id to the session
				//variable named "id" so it can be reused in the later steps
				//The path extraction is using https://github.com/tidwall/gjson notation
				pitcher.Extract("id", "0.id"),
			),

		//Manually creating the step and the request
		pitcher.Step{
			Request: &pitcher.Request{
				Method: "GET",

				// Path reusing the "id" variable extracted in the previous steps
				Path: "/posts/${id}",
			},
			Assertions: []pitcher.AssertionFunc{
				pitcher.SuccessAssertion,
			},
		},

		//pitcher.POST("<url>",<body>, <contentType>)
		pitcher.POST(
			"/posts",
			`{"title": "Michael G Scott", "body": "Regional Manager ${randomUUID}", "userId": 1 }`,
			"application/json",
		).WithPreProcessors(
			pitcher.JWTAuth,
		).WithPostProcessors(
			//Post processor extracting the id again and overriding the previous value
			pitcher.Extract("id", "id"),

			//Post processor logging the entire payload response
			pitcher.LogPayloadProcessor,
		),
		pitcher.Step{
			Request: &pitcher.Request{
				Method: "GET",
				Path:   "/post/${id}",
			},
			Assertions: []pitcher.AssertionFunc{
				pitcher.NotFoundAssertion,
			},
		},
	)

	if err != nil {
		panic(err)
	}
}

func parametersFromEnv() map[string]string {
	envVar := flag.String("env", "local", "defines environment name to be used to run the calls")
	flag.Parse()

	env := local

	switch *envVar {
	case "dev":
		fmt.Println("Using Dev environment")
		env = dev
	}
	return env
}
