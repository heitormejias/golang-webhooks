Library webhooks
================


Library webhooks allows for easy receiving and parsing of Git Webhook Events

Gits Support:
* GitHub
* Bitbucket
* GitLab
* Gogs (WIP)
* Gitea (WIP)
* Docker

Features:

* Parses the entire payload, not just a few fields.
* Fields + Schema directly lines up with webhook posted json

Notes:

* Currently only accepting json payloads.

Installation
------------

Use go get.

```shell
go get -u github.com/heitormejias/golang-webhooks
```

Then import the package into your own code.

	import "github.com/heitormejias/golang-webhooks"

Usage and Documentation
------

Please see http://godoc.org/github.com/heitormejias/golang-webhooks for detailed usage docs.

##### Examples:

See more on "_examples" folder

```go
package main

import (
	"fmt"

	"net/http"

	"github.com/heitormejias/golang-webhooks/github"
)

func main() {
	hook, _ := github.New(github.Options.Secret("MyGitHubSuperSecretSecrect...?"))

	// path: "/github/webhooks"
	http.HandleFunc(github.ServerPath, func(w http.ResponseWriter, r *http.Request) {
		payload, err := hook.Parse(r, github.ReleaseEvent, github.PullRequestEvent)
		if err != nil {
			if err == github.ErrEventNotFound {
				// ok event wasn;t one of the ones asked to be parsed
			}
		}
		switch payload.(type) {

		case github.ReleasePayload:
			release := payload.(github.ReleasePayload)
			// Do whatever you want from here...
			fmt.Printf("%+v", release)

		case github.PullRequestPayload:
			pullRequest := payload.(github.PullRequestPayload)
			// Do whatever you want from here...
			fmt.Printf("%+v", pullRequest)
		}
	})
	http.ListenAndServe(":3000", nil)
}

```

Contributing
------

Pull requests for other services are welcome!

If the changes being proposed or requested are breaking changes, please create an issue for discussion.

License
------
Distributed under MIT License, please see license file in code for more details.
