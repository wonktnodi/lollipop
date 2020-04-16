package main

import (
  "fmt"
  "io/ioutil"
  "log"
  "net/http"
  "net/http/httptest"
  "net/http/httputil"
  "net/url"
  "strings"
)

func main() {
  //testFunc2()
  testFunc3()
}

type ReverseProxy struct {
  httputil.ReverseProxy

  currentDirector int
  Directors       []func(req *http.Request)
}

func NewMultiHostReverseProxy(targets ...*url.URL) (reverseProxy *ReverseProxy) {
  // grab a new ReverseProxy and set the currentDirector to 0
  reverseProxy = &ReverseProxy{currentDirector: 0}

  // loop through given targets and add as Directors
  for _, target := range targets {
    reverseProxy.AddHost(target)
  }

  // send back the
  return reverseProxy
}

func singleJoiningSlash(a, b string) string {
  aslash := strings.HasSuffix(a, "/")
  bslash := strings.HasPrefix(b, "/")
  switch {
  case aslash && bslash:
    return a + b[1:]
  case !aslash && !bslash:
    return a + "/" + b
  }
  return a + b
}

func (p *ReverseProxy) AddHost(target *url.URL) {
  targetQuery := target.RawQuery
  p.Directors = append(p.Directors, func(req *http.Request) {
    req.URL.Scheme = target.Scheme
    req.URL.Host = target.Host
    req.URL.Path = singleJoiningSlash(target.Path, req.URL.Path)
    if targetQuery == "" || req.URL.RawQuery == "" {
      req.URL.RawQuery = targetQuery + req.URL.RawQuery
    } else {
      req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
    }
    if _, ok := req.Header["User-Agent"]; !ok {
      // explicitly disable User-Agent so it's not set to default value
      req.Header.Set("User-Agent", "")
    }
  })
}

func (p *ReverseProxy) NextDirector() func(*http.Request) {
  log.Printf("Current Director: %d", p.currentDirector)
  p.currentDirector += 1
  log.Printf("Next Director: %d", p.currentDirector)
  if p.currentDirector >= len(p.Directors) {
    p.currentDirector = 0
    log.Printf("Too High; Resetting Director: %d", p.currentDirector)
  }

  log.Printf("Returning Next Director: %d", p.currentDirector)
  return p.Directors[p.currentDirector]
}

func (p *ReverseProxy) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
  // Rotate Director
  p.Director = p.NextDirector()

  // Call the upstream ServeHTTP
  p.ReverseProxy.ServeHTTP(rw, req)
}

func testFunc3() {
  var targets []*url.URL

  for _, host := range []string{"http://httpbin.org"} {
    rpURL, err := url.Parse(host)
    if err != nil {
      log.Fatalf("Error parsing URL: %v", err)
    }
    targets = append(targets, rpURL)
  }

  rp := NewMultiHostReverseProxy(targets...)

  http.HandleFunc("/", rp.ServeHTTP)

  log.Fatal(http.ListenAndServe(":8888", nil))

  //feProxy := &http.Server{
  //  Addr:    ":80",
  //  Handler: rp,
  //}
  //defer func() {
  //  err := feProxy.Close()
  //  if err != nil {
  //    log.Fatalf("Error closing proxy: %v", err)
  //  }
  //}()
  //
  //err := feProxy.ListenAndServe()
  //if err != nil {
  //  log.Fatal("Error Listening: ", err)
  //}
}

func testFunc2() {
  backendServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintln(w, "this call was relayed by the reverse proxy")
  }))
  defer backendServer.Close()

  rpURL, err := url.Parse(backendServer.URL)
  if err != nil {
    log.Fatal(err)
  }
  frontendProxy := httptest.NewServer(httputil.NewSingleHostReverseProxy(rpURL))
  defer frontendProxy.Close()

  resp, err := http.Get(frontendProxy.URL)
  if err != nil {
    log.Fatal(err)
  }

  b, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    log.Fatal(err)
  }

  fmt.Printf("%s", b)
}

func testFunc1() {
  http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    director := func(req *http.Request) {
      req = r
      req.URL.Scheme = "http"
      req.URL.Host = r.Host
    }
    proxy := &httputil.ReverseProxy{Director: director}
    proxy.ServeHTTP(w, r)
  })

  log.Fatal(http.ListenAndServe(":8888", nil))
}
