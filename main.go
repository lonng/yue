package main

import (
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/chrislonng/nex"
	"github.com/gorilla/mux"
)

var ErrInvalidParameter = errors.New("invalid parameter")

var token = "123456"

// save clues
var db = struct{ clues []ClueInfo }{}

func main() {
	nex.SetErrorEncoder(func(err error) interface{} {
		return &ErrorMessage{
			Code:  -1000,
			Error: err.Error(),
		}
	})

	r := mux.NewRouter()

	r.Handle("/clues", nex.Handler(createClue)).Methods("POST")
	r.Handle("/clues", nex.Handler(clueList)).Methods("GET")

	r.Handle("/clues/{id}", nex.Handler(clueInfo)).Methods("GET")
	r.Handle("/clues/{id}", nex.Handler(updateClue)).Methods("PUT")
	r.Handle("/clues/{id}", nex.Handler(deleteClue)).Methods("DELETE")

	r.Handle("/blob", nex.Handler(uploadFile)).Methods("POST")

	if err := http.ListenAndServe(":8080", r); err != nil {
		panic(err)
	}
}

func createClue(c *ClueInfo) (*StringMessage, error) {
	title := strings.TrimSpace(c.Title)
	number := strings.TrimSpace(c.Number)

	if title == "" || number == "" {
		return nil, errors.New("title and number can not empty")
	}

	db.clues = append(db.clues, *c)
	return SuccessResponse, nil
}

func clueList(query nex.Form) (*ClueListResponse, error) {
	s := query.Get("start")
	c := query.Get("count")

	var start, count int
	var err error

	if s == "" {
		start = 0
	} else {
		start, err = strconv.Atoi(s)
		if err != nil {
			return nil, err
		}
	}

	if c == "" {
		count = len(db.clues)
	} else {
		count, err = strconv.Atoi(c)
		if err != nil {
			return nil, err
		}
	}

	return &ClueListResponse{Data: db.clues[start : start+count]}, nil
}

// util function
func parseID(r *http.Request) (int, error) {
	vars := mux.Vars(r)
	if strings.TrimSpace(vars["id"]) == "" {
		return 0, ErrInvalidParameter
	}

	id, err := strconv.Atoi(strings.TrimSpace(vars["id"]))
	if err != nil {
		return 0, err
	}

	if len(db.clues) <= id-1 {
		return 0, errors.New("can not found clue information")
	}

	return id, nil
}

func clueInfo(r *http.Request) (*ClueInfoResponse, error) {
	id, err := parseID(r)
	if err != nil {
		return nil, err
	}

	return &ClueInfoResponse{Data:&db.clues[id-1]}, nil
}

func updateClue(r *http.Request, c *ClueInfo) (*StringMessage, error) {
	id, err := parseID(r)
	if err != nil {
		return nil, err
	}

	title := strings.TrimSpace(c.Title)
	number := strings.TrimSpace(c.Number)

	if title == "" || number == "" {
		return nil, errors.New("title and number can not empty")
	}

	db.clues[id] = *c

	return SuccessResponse, nil
}

func deleteClue(h http.Header, r *http.Request) (*StringMessage, error) {
	t := h.Get("Authorization")
	if t != token {
		return nil, errors.New("permission denied")
	}

	id, err := parseID(r)
	if err != nil {
		return nil, err
	}

	db.clues = append(db.clues[:id], db.clues[id:]...)
	return SuccessResponse, nil
}

func uploadFile(form *multipart.Form) (*BlobResponse, error) {
	uploaded, ok := form.File["uploadfile"]
	if !ok {
		return nil, errors.New("can not found `uploadfile` field")
	}

	localName := func(filename string) string {
		ext := filepath.Ext(filename)
		id := time.Now().Format("20060102150405.999999999")
		return id + ext
	}

	var fds []io.Closer
	defer func() {
		for _, fd := range fds {
			fd.Close()
		}
	}()

	files := make(map[string]string)
	for _, fh := range uploaded {
		fileName := localName(fh.Filename)
		files[fh.Filename] = fileName

		// upload file
		uf, err := fh.Open()
		fds = append(fds, uf)
		if err != nil {
			return nil, err
		}

		// local file
		lf, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY, 0660)
		fds = append(fds, lf)
		if err != nil {
			return nil, err
		}

		_, err = io.Copy(lf, uf)
		if err != nil {
			return nil, err
		}
	}

	return &BlobResponse{Data: files}, nil
}
