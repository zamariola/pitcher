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

	err := client.Do(
		pitcher.Step{

			//Request definition (method, path, payload, headers)
			Request: &pitcher.Request{
				Method: "GET",
				Path:   "/posts",
			},

			//Assertions after execution
			Assertions: []pitcher.AssertionFunc{
				pitcher.SuccessAssertion,
			},

			PostProcs: []pitcher.PostProcessorFunc{

				//Post processor that extracts the json value from the path 0.id to the session
				//variable named "id" so it can be reused in the later steps
				//The path extraction is using https://github.com/tidwall/gjson notation
				pitcher.Extract("id", "0.id"),
			},
		},
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
		pitcher.Step{

			//Post request using the body
			//The content-type: application/json is the default if no other is informed
			Request: &pitcher.Request{
				Method: "POST",
				Path:   "/posts",
				Body:   `{"title": "Michael G Scott", "body": "Regional Manager ${randomUUID}", "userId": 1 }`,
			},
			Assertions: []pitcher.AssertionFunc{
				pitcher.SuccessAssertion,
			},
			PreProcs: []pitcher.PreProcessorFunc{},
			PostProcs: []pitcher.PostProcessorFunc{

				//Post processor extracting the id again and overriding the previous value
				pitcher.Extract("id", "id"),
				pitcher.LogPayloadProcessor,
			},
		},
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
