// This module ensures that you embed the current git revision into the
// application binary. You can do so with the following command:
//
//     go build -ldflags "-X main.gitRevisionId $(git rev-parse HEAD)"

package main

import (
	"fmt"
	"net/http"
)

var gitRevisionId string

func init() {
	if gitRevisionId == "" {
		panic(fmt.Errorf("you must define a version number during compilation"))
	}
}

func versionServer(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, fmt.Sprintf("https://github.com/sburnett/encore/commit/%[1]s", gitRevisionId), http.StatusFound)
	fmt.Fprint(w, gitRevisionId)
}
