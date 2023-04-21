package cmd

import (
	"fmt"
	"http-attenuator/data"
	"io"
	"log"
	"net"
	"net/http"
	"strings"

	"github.com/elazarl/goproxy"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var proxyCmd = &cobra.Command{
	Use:   "proxy",
	Short: "Runs the API proxy",
	Run:   RunProxy,
}

var proxyAddress string

func init() {
	proxyCmd.PersistentFlags().StringVarP(&proxyAddress, "listen", "l", "0.0.0.0:8888", "API listen address (default is 0.0.0.0:8002)")
}

type HandleAllRequests struct {
}

func (h *HandleAllRequests) HandleResp(resp *http.Response, ctx *goproxy.ProxyCtx) bool {
	return true
}

func RunProxy(cmd *cobra.Command, args []string) {
	if viper.GetString(data.CONF_PROXY_LISTEN) != "" {
		proxyAddress = viper.GetString(data.CONF_PROXY_LISTEN)
	}

	// Enable metrics
	go RunMetrics(cmd, args)

	handler := &proxy{}

	log.Printf("Running API proxy on %s", proxyAddress)
	if err := http.ListenAndServe(proxyAddress, handler); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

// Hop-by-hop headers. These are removed when sent to the backend.
// http://www.w3.org/Protocols/rfc2616/rfc2616-sec13.html
var hopHeaders = []string{
	"Connection",
	"Keep-Alive",
	"Proxy-Authenticate",
	"Proxy-Authorization",
	"Te", // canonicalized version of "TE"
	"Trailers",
	"Transfer-Encoding",
	"Upgrade",
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func delHopHeaders(header http.Header) {
	for _, h := range hopHeaders {
		header.Del(h)
	}
}

func appendHostToXForwardHeader(header http.Header, host string) {
	// If we aren't the first proxy retain prior
	// X-Forwarded-For information as a comma+space
	// separated list and fold multiple headers into one.
	if prior, ok := header["X-Forwarded-For"]; ok {
		host = strings.Join(prior, ", ") + ", " + host
	}
	header.Set("X-Forwarded-For", host)
}

type proxy struct {
}

func (p *proxy) ServeHTTP(wr http.ResponseWriter, req *http.Request) {
	log.Println(req.RemoteAddr, " ", req.Method, " ", req.URL)
	proxyRequests.WithLabelValues(req.Host).Inc()

	if req.URL.Scheme != "http" && req.URL.Scheme != "https" {
		msg := "unsupported protocal scheme " + req.URL.Scheme
		http.Error(wr, msg, http.StatusBadRequest)
		log.Println(msg)
		return
	}

	client := &http.Client{}

	//http: Request.RequestURI can't be set in client requests.
	//http://golang.org/src/pkg/net/http/client.go
	req.RequestURI = ""

	delHopHeaders(req.Header)

	if clientIP, _, err := net.SplitHostPort(req.RemoteAddr); err == nil {
		appendHostToXForwardHeader(req.Header, clientIP)
	}

	resp, err := client.Do(req)
	// TODO(john): deal with 'bad gateway'
	if err != nil {
		http.Error(wr, "Server Error", http.StatusInternalServerError)
		log.Fatal("ServeHTTP:", err)
	}
	defer resp.Body.Close()
	proxyResponses.WithLabelValues(req.Host, fmt.Sprint(resp.StatusCode)).Inc()

	log.Println(req.RemoteAddr, " ", resp.Status)

	delHopHeaders(resp.Header)

	copyHeader(wr.Header(), resp.Header)
	wr.WriteHeader(resp.StatusCode)
	io.Copy(wr, resp.Body)
}
