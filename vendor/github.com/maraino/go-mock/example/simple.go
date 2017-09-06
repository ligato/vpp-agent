// +build ignore

// Simple application to show the use of go-mock.
package main

import (
	"fmt"
	"net/url"

	"github.com/maraino/go-mock"
)

type Client interface {
	Request(url *url.URL) (int, string, error)
}

type MyClient struct {
	mock.Mock
}

func (c *MyClient) Request(url *url.URL) (int, string, error) {
	ret := c.Called(url)
	return ret.Int(0), ret.String(1), ret.Error(2)
}

func main() {
	c := &MyClient{}

	url, _ := url.Parse("http://www.example.org")
	c.When("Request", url).Return(200, "{result:1}", nil).Times(1)
	c.When("Request", mock.Any).Return(500, "{result:0}", fmt.Errorf("Internal Server Error")).Times(1)

	code, json, err := c.Request(url)
	fmt.Printf("Code: %d, JSON: %s, Error: %v\n", code, json, err)

	url, _ = url.Parse("http://www.github.com")
	code, json, err = c.Request(url)
	fmt.Printf("Code: %d, JSON: %s, Error: %v\n", code, json, err)

	if ok, err := c.Verify(); !ok {
		fmt.Println(err)
	}
}
