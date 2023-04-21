package api

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
)

func BrokerHandler(c *gin.Context) {
	// Extract the host from the URL
	hostAndQuery := c.Param("hostAndQuery")
	if hostAndQuery == "" {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	for hostAndQuery[0:1] == "/" {
		hostAndQuery = hostAndQuery[1:]
	}
	log.Printf("%+v", hostAndQuery)

	hostAndQueryUrl, err := url.Parse(hostAndQuery)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	request := *c.Request
	request.URL = hostAndQueryUrl
	request.Host = hostAndQueryUrl.Host

	//http: Request.RequestURI can't be set in client requests.
	//http://golang.org/src/pkg/net/http/client.go
	request.RequestURI = ""

	// Make the request
	//
	// TODO(john): put it through the attenuator / circuit breaker etc
	client := http.Client{}
	rb, err := json.MarshalIndent(request, "", "  ")
	if err != nil {
		log.Println(err.Error())
	}
	log.Println(string(rb))
	resp, err := client.Do(&request)
	if err != nil {
		log.Printf("%s: %s", hostAndQuery, err.Error())
		c.AbortWithError(http.StatusBadGateway, err)
		return
	}

	defer resp.Body.Close()

	// Send the status
	c.Status(resp.StatusCode)

	// Send the headers
	for h, v := range resp.Header {
		for _, headerVal := range v {
			c.Writer.Header().Add(h, headerVal)
		}
	}

	// Send the body
	io.Copy(c.Writer, resp.Body)
}
