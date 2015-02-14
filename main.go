package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

//Response represents a queued response
type Response struct {
	Status  int
	Headers map[string]string
	Body    string
	Wait    int64
}

//SessionStore stores all enqueued responses
type SessionStore struct {
	sessions map[string][]Response
	lock     sync.RWMutex
}

//Enqueue a response for a named session
func (s *SessionStore) Enqueue(sessionName string, response Response) int {
	s.lock.Lock()
	if s.sessions == nil {
		s.sessions = make(map[string][]Response)
	}
	s.sessions[sessionName] = append(s.sessions[sessionName], response)

	length := len(s.sessions[sessionName])
	s.lock.Unlock()

	return length
}

//Dequeue a response from a named session (FIFO)
func (s *SessionStore) Dequeue(sessionName string) Response {

	if s.sessions != nil {
		s.lock.Lock()
		//check the session exists and has responses
		if _, ok := s.sessions[sessionName]; ok && len(s.sessions[sessionName]) > 0 {
			var itm Response
			itm, s.sessions[sessionName] = s.sessions[sessionName][0], s.sessions[sessionName][1:]
			s.lock.Unlock()
			return itm
		}
		s.lock.Unlock()
	}

	return Response{}
}

func main() {

	sessions := SessionStore{}

	routes := mux.NewRouter()

	//record sessions
	routes.HandleFunc("/r/{sess}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)

		decoder := json.NewDecoder(r.Body)
		var response Response
		if err := decoder.Decode(&response); err != nil {
			w.WriteHeader(400)
			fmt.Fprintf(w, "Bad Request: %s", err)
			log.Print("error: request did not contain valid JSON")
			return
		}

		log.Printf("added response to %v", vars["sess"])
		fmt.Fprintf(w, fmt.Sprintf("%v responses in session", sessions.Enqueue(vars["sess"], response)))

	}).Methods("POST")

	//serve sessions
	routes.HandleFunc("/p/{sess}/{path:.*}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		res := sessions.Dequeue(vars["sess"])

		//wait time
		if res.Wait > 0 {
			time.Sleep(time.Duration(res.Wait) * time.Millisecond)
		}

		//headers
		w.WriteHeader(res.Status)
		for name, val := range res.Headers {
			w.Header().Set(name, val)
		}

		//body
		fmt.Fprint(w, res.Body)

		log.Printf("returned response for %s", vars["sess"])
	})

	var port = flag.String("port", "8080", "port to listen on")
	flag.Parse()

	log.Printf("Starting server on port %v", *port)
	http.ListenAndServe(":"+*port, routes)
}
