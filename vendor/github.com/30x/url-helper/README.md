# Url Helper

Small library to help generate absolute url paths based on incoming request. Uses `X-Forwarded-Host` and `X-Forwarded-Proto` first to gather Host and HTTP Scheme used generated urls.

```golang

package main

import (
	"fmt"
	"github.com/30x/url-helper"
	"net/http"
)

func main() {

	http.HandleFunc("/some/resource", func(w http.ResponseWriter, r *http.Request) {
		url, _ := urlhelper.NewURLHelper(r)
		fmt.Println(url.Current())                            // GET http://1.2.3.4/some/resource?test=123 -> http://1.2.3.4/some/resource?test=123
		fmt.Println(url.Join("v1"))                           // GET http://1.2.3.4/some/resource?test=123 -> http://1.2.3.4/some/resource/v1?test=123
		fmt.Println(url.Join("../other"))                     // GET http://1.2.3.4/some/resource?test=123 -> http://1.2.3.4/some/other?test=123
		fmt.Println(url.Path("/new/root"))                    // GET http://1.2.3.4/some/resource?test=123 -> http://1.2.3.4/new/root
		fmt.Println(url.PathWithQuery("/new/root", "page=1")) // GET http://1.2.3.4/some/resource?test=123 -> http://1.2.3.4/new/root?page=1
	})

	http.ListenAndServe(":8080", nil)
}

```
