package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// getReq is the request that is sent to the server to get a quote of the day.
type getReq struct {
	// Author is the author whose quote you want. If left empty, will randomly choose
	// an author.
	Author string `json:"author"`
}

// fromReader reads from an io.Reader and unmarshal's the content into the getReq{}. This
// is used to decode from the http.Request.Body into our struct.
func (g *getReq) fromReader(r io.Reader) error {
	b, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	return json.Unmarshal(b, g)
}

// getResp is the response from the server for a quote of the day.
type getResp struct {
	// Quote is the quote from the server.
	Quote string `json:"quote"`
	// Error is an error if the server had a non-http related error.
	Error *Error `json:"error"`
}

// ErrCode is an error code so the user can tell what the specific error condition was.
type ErrCode string

// Error is a custom error type for this package.
type Error struct {
	// Code is the error code.
	Code ErrCode
	// Msg is the textual message.
	Msg string
}

// Error implements error.Error().
func (e Error) Error() string {
	return fmt.Sprintf("(code %v): %s", e.Code, e.Msg)
}

const (
	// UnknownCode indicates the ErrCode was not set, aka the zero value.
	UnknownCode ErrCode = ""
	// UnknownAuthor indicates that the request wanted a quote from an
	// author that didn't exist in the server.
	UnknownAuthor ErrCode = "UnknownAuthor"
)

/*
______ _____ _____ _____   _____ _ _            _
| ___ \  ___/  ___|_   _| /  __ \ (_)          | |
| |_/ / |__ \ `--.  | |   | /  \/ |_  ___ _ __ | |_
|    /|  __| `--. \ | |   | |   | | |/ _ \ '_ \| __|
| |\ \| |___/\__/ / | |   | \__/\ | |  __/ | | | |_
\_| \_\____/\____/  \_/    \____/_|_|\___|_| |_|\__|

*/

// QOTD represents our client to talk to the QOTD server.
type QOTD struct {
	// u is the URL for the server's address, aka http://someserver.com:80
	u *url.URL
	// client is the *http.Client that will be reused to contact the server.
	client *http.Client
}

// New constructs a new QOTD client.
func New(addr string) (*QOTD, error) {
	u, err := url.Parse(addr)
	if err != nil {
		return nil, err
	}

	return &QOTD{
		u:      u,
		client: &http.Client{},
	}, nil
}

// restCall provides a generic POST and JSON REST call function. This can be reused by future
// calls to other REST endpoints.
func (q *QOTD) restCall(ctx context.Context, endpoint string, req, resp interface{}) error {
	// If we don't have a deadline, apply a default.
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 2*time.Second)
		defer cancel()
	}

	// Conert our request into JSON.
	b, err := json.Marshal(req)
	if err != nil {
		return err
	}

	// Create a new HTTP request using POST to out endpoint with the body
	// set to our JSON request.
	hReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		endpoint,
		bytes.NewBuffer(b),
	)
	if err != nil {
		return err
	}

	// Make the request.
	hResp, err := q.client.Do(hReq)
	if err != nil {
		return err
	}

	// Read the response's Body.
	b, err = io.ReadAll(hResp.Body)
	if err != nil {
		return err
	}

	// Unmarshal the JSON into the response.
	return json.Unmarshal(b, resp)
}

// Get fetches a qutoe of the day from the server. If the author is not set, a random one is chosen.
func (q *QOTD) Get(ctx context.Context, author string) (string, error) {
	const endpoint = `/qotd/v1/get`
	ref, _ := url.Parse(endpoint)

	resp := getResp{}

	// Makes a call to the server. The endpoint is the joining of our base url (http://127.0.0.1:80) with
	// our endpoint constant above (`qotd/v1/get`) to form `http://127.0.0.1:80/qotd/v1/get`.
	err := q.restCall(ctx, q.u.ResolveReference(ref).String(), getReq{Author: author}, &resp)
	switch {
	case err != nil: // http error
		return "", err
	case resp.Error != nil: // server error, such as the author not being found.
		return "", resp.Error
	}

	return resp.Quote, nil
}

/*
______ _____ _____ _____   _____
| ___ \  ___/  ___|_   _| /  ___|
| |_/ / |__ \ `--.  | |   \ `--.  ___ _ ____   _____ _ __
|    /|  __| `--. \ | |    `--. \/ _ \ '__\ \ / / _ \ '__|
| |\ \| |___/\__/ / | |   /\__/ /  __/ |   \ V /  __/ |
\_| \_\____/\____/  \_/   \____/ \___|_|    \_/ \___|_|

*/

// server is a REST server for serving quotes of the day.
type server struct {
	// serv is the http server we will use.
	serv *http.Server
	// quotes has keys that are names and values that are list of quotes attributed
	// to that person.
	quotes map[string][]string
}

// newServer is the constructor for server. The port is the port to run on.
func newServer(port int) (*server, error) {
	s := &server{
		serv: &http.Server{
			Addr: ":" + strconv.Itoa(port), // results in a string like: ":80"
		},
		quotes: map[string][]string{
			"Mark Twain": {
				"History doesn't repeat itself, but it does rhyme",
				"Lies, damned lies, and statistics",
				"Golf is a good walk spoiled",
			},
			"Benjamin Franklin": {
				"Tell me and I forget. Teach me and I remember. Involve me and I learn",
				"I didn't fail the test. I just found 100 ways to do it wrong",
			},
			"Eleanor Roosevelt": {
				"The future belongs to those who believe in the beauty of their dreams",
			},
		},
	}
	// A mux handles looking at an incoming URL and determining what function should handle it.
	// This has rules for pattern matching you can read more about here: https://pkg.go.dev/net/http#ServeMux
	mux := http.NewServeMux()
	mux.HandleFunc(`/qotd/v1/get`, s.qotdGet)

	// The muxer implements http.Handler and we assign it for our servers URL handling.
	s.serv.Handler = mux

	return s, nil
}

// start starts our server.
func (s *server) start() error {
	return s.serv.ListenAndServe()
}

// qotdGet provides an http.HandleFunc for receiving REST requests for a quote of the day.
func (s *server) qotdGet(w http.ResponseWriter, r *http.Request) {
	// Get the Context for the request.
	ctx := r.Context()

	// If no deadline is set, set one.
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 2*time.Second)
		defer cancel()
	}

	// read our http.Request's body as JSON into our request object.
	req := getReq{}
	if err := req.fromReader(r.Body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var quotes []string

	// No author was requested so we will randomly choose one.
	if req.Author == "" {
		// To get a value from a map, you must know the key.
		// Since we are trying to get a quote from a random author,
		// we will simply do a single loop using range that extracts
		// from the map in random order.
		for _, quotes = range s.quotes {
			break
		}
	} else { // Author was requested.
		// Find the autors.
		var ok bool
		quotes, ok = s.quotes[req.Author]
		// Not author was found, send a custom error message back.
		if !ok {
			b, err := json.Marshal(
				getResp{
					Error: &Error{
						Code: UnknownAuthor,
						Msg:  fmt.Sprintf("Author %q was not found", req.Author),
					},
				},
			)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			w.Write(b)
			return
		}
	}

	// This chooses a random number whose maximum value is the length of our quotes slice.
	// Note that `math/rand` calls vs `crypto/rand` calls are not cryptographically secure.
	i := rand.Intn(len(quotes))

	// Send our quote back to the client.
	b, err := json.Marshal(getResp{Quote: quotes[i]})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Write(b)
	return
}
