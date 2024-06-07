# Pitcher

![Pitcher Logo](https://github.com/zamariola/pitcher/blob/eb35ebd720043189e30fb0002354ab79f688f845/docs/pitcher-mascot.png)

Pitcher is an open-source API caller tool designed to simplify the process of making HTTP requests and testing APIs. Similar to Postman, it provides a user-friendly programatic interface for crafting requests, inspecting responses, and organizing your API workflow.

It's Golang code centric, with no UI, no versioning and no copy and pasting. The real developer productivity tool.

## Features

- **User-Friendly Interface**: Intuitive UI makes it easy to create and send HTTP requests.
- **Multiple Request Methods**: Support for various HTTP methods including GET, POST, PUT, DELETE, etc.
- **Environment Variables**: Define and utilize environment variables for flexible testing.
- **Authentication**: Support for different authentication methods including Basic Auth, OAuth, JWT, etc.
- **Logs**: Support for different authentication methods including Basic Auth, OAuth, JWT, etc.

## Installation

- Option1: To install Pitcher, simply clone the repository and update the "cmd/main.go" to your needs or
- Option2: Using go get
```go get github.com/zamariola/pitcher```

and the create the project importing it

```go
package main

import (
	"flag"
	"fmt"

	pitcher "github.com/zamariola/pitcher"
)

var local = map[string]string{
	"host": "https://jsonplaceholder.typicode.com",
}

func main() {

	// Define the client with the global processors
	client := pitcher.NewClientWithProcessors(

		//Read and write session that expires once it finishes
		pitcher.NewMemoryRWSession(local),

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

```


## Contributing

We welcome contributions from the community! If you'd like to contribute to Pitcher, please follow these steps:

1. Fork the repository.
2. Create a new branch (`git checkout -b feature/improvement`).
3. Make your changes.
4. Commit your changes (`git commit -am 'Add new feature'`).
5. Push to the branch (`git push origin feature/improvement`).
6. Create a new Pull Request.

Please make sure to read our [Contribution Guidelines](link-to-contribution-guidelines) before submitting your pull request.

## Support

If you encounter any issues or have any questions about Pitcher, feel free to [open an issue](https://github.com/zamariola/pitcher/issues) on GitHub.

## License

Pitcher is licensed under the [MIT License](link-to-license).

## Acknowledgements

Pitcher wouldn't be possible without the contributions of the following individuals:

- [Leonardo Zamariola](https://github.com/zamariola): Maintainer.