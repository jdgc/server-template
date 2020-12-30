package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"jdgc/lists-server/v2/cmd"
	"jdgc/lists-server/v2/db"
	"net/http"
	"strings"
	"sync"
	"time"
)

type List struct {
	ID    string     `json:"id"`
	Name  string     `json:"name"`
	Items []ListItem `json:"items"`
}

type ListItem struct {
	ID   string `json:"id"`
	Data string `json:"data"`
}

type listHandlers struct {
	sync.Mutex
	store map[string]List
}

func newListHandlers() *listHandlers {
	// store should be from DB once thats .in
	return &listHandlers{
		store: map[string]List{},
	}
}

func main() {
	listHandlers := newListHandlers()
	db.Init()

	cmd.Execute()

	http.HandleFunc("/lists", listHandlers.lists)
	http.HandleFunc("lists/", listHandlers.getList)

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
	fmt.Println("Server running on port 8080")
}

func (h *listHandlers) lists(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		h.get(w, r)
		return
	case "POST":
		h.post(w, r)
		return
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("method not allowed"))
		return
	}
}

func (h *listHandlers) post(w http.ResponseWriter, r *http.Request) {
	bodyBytes, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}

	contentType := r.Header.Get("content-type")
	if contentType != "application/json" {
		w.WriteHeader(http.StatusUnsupportedMediaType)
		w.Write([]byte(fmt.Sprintf(
			"require content-type 'application-json', but got '%s'", contentType),
		))
		return
	}

	var list List
	err = json.Unmarshal(bodyBytes, &list)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	list.ID = fmt.Sprintf("%d", time.Now().UnixNano())

	h.Lock()
	h.store[list.ID] = list
	defer h.Unlock()
}

func (h *listHandlers) get(w http.ResponseWriter, r *http.Request) {
	lists := make([]List, len(h.store))

	// init read mutex lock
	h.Lock()
	i := 0
	for _, list := range h.store {
		lists[i] = list
		i++
	}
	h.Unlock()

	jsonBytes, err := json.Marshal(lists)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}

	w.Header().Add("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
}

func (h *listHandlers) getList(w http.ResponseWriter, r *http.Request) {
	urlParts := strings.Split(r.URL.String(), "/")

	if len(urlParts) != 3 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// replace this with select by ID SQL
	h.Lock()
	list, ok := h.store[urlParts[2]]
	h.Unlock()

	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	jsonBytes, err := json.Marshal(list)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}

	w.Header().Add("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
}
