package main

import (
	"bytes"
	"flag"
	"github.com/pranavraja/front/cache"
	"io/ioutil"
	"net/http"
	"time"
)

var upstream string
var host string
var TTLSeconds int64

func getUpstream(path string) (data []byte, ttl time.Duration) {
	resp, err := http.Get("http://" + upstream + path)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	buf = bytes.Replace(buf, []byte(upstream), []byte(host), -1)
	return buf, time.Duration(TTLSeconds) * time.Second
}

func init() {
	flag.StringVar(&upstream, "upstream", "", "The hostname of the upstream server")
	flag.StringVar(&host, "host", "localhost:8080", "The current server's host:port (used to rewrite URLs in the response to point to the proxy instead of the origin)")
	flag.Int64Var(&TTLSeconds, "ttl", 3600, "The assumed TTL of responses from the upstream server, in seconds (default 1 hour)")
}

func main() {
	flag.Parse()
	if upstream == "" {
		println("Invalid upstream host: " + upstream)
		return
	}
	c := cache.New(getUpstream)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		buf, _ := c.Get(path)
		if buf == nil {
			http.Error(w, "Couldn't fetch "+upstream+path, http.StatusInternalServerError)
			return
		}
		_, err := w.Write(buf)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
	http.ListenAndServe(":8080", nil)
}
