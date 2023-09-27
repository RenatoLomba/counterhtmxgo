package main

import (
	"log"
	"net/http"
	"sync"
	"text/template"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
)

type Counter struct {
	value int
	mu    sync.Mutex
}

func (c *Counter) Increase() {
	c.mu.Lock()
	c.value++
	c.mu.Unlock()
}

func (c *Counter) Decrease() {
	c.mu.Lock()
	c.value--
	c.mu.Unlock()
}

func (c *Counter) GetValue() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.value
}

var sessionManager *scs.SessionManager

func main() {
	log.Println("Starting server on port 3000...")

	sessionManager = scs.New()
	sessionManager.Lifetime = 24 * time.Hour

	db := map[string]*Counter{}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl := template.Must(template.ParseFiles("index.html"))

		userId := sessionManager.GetString(r.Context(), "user-id")
		if userId == "" {
			userId = uuid.New().String()
			sessionManager.Put(r.Context(), "user-id", userId)
		}

		userCounter := db[userId]
		if userCounter == nil {
			userCounter = &Counter{}
			db[userId] = userCounter
		}

		data := map[string]int{
			"CounterValue": userCounter.GetValue(),
		}

		tmpl.Execute(w, data)
	})
	r.Post("/increase", func(w http.ResponseWriter, r *http.Request) {
		tmplStr := "<span id=\"counter\">{{.CounterValue}}</span>"
		tmpl := template.Must(template.New("counter").Parse(tmplStr))

		userId := sessionManager.GetString(r.Context(), "user-id")
		userCounter := db[userId]

		userCounter.Increase()
		data := map[string]int{
			"CounterValue": userCounter.GetValue(),
		}

		tmpl.ExecuteTemplate(w, "counter", data)
	})
	r.Post("/decrease", func(w http.ResponseWriter, r *http.Request) {
		tmplStr := "<span id=\"counter\">{{.CounterValue}}</span>"
		tmpl := template.Must(template.New("counter").Parse(tmplStr))

		userId := sessionManager.GetString(r.Context(), "user-id")
		userCounter := db[userId]

		userCounter.Decrease()
		data := map[string]int{
			"CounterValue": userCounter.GetValue(),
		}

		tmpl.ExecuteTemplate(w, "counter", data)
	})

	log.Fatal(http.ListenAndServe(":3000", sessionManager.LoadAndSave(r)))
}
