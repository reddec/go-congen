// Code generated by go-congen DO NOT EDIT.
package controller

import (
	_ "embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
)

//go:embed index.html
var templateContent string

type Controller interface {
	ViewData(request *http.Request, lastError error) (interface{}, error)
	Do(writer http.ResponseWriter, request *http.Request, params Params) error
	DoDelete(writer http.ResponseWriter, request *http.Request, params DeleteParams) error
	DoResetPassword(writer http.ResponseWriter, request *http.Request, params ResetPasswordParams) error
}

func Wrap(controller Controller) http.Handler {
	t := template.Must(template.New("").Parse(templateContent))
	w := &wrapper{template: t, controller: controller}
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		switch request.Method {
		case http.MethodGet:
			w.mainView(writer, request, nil)
		case http.MethodPost:
			w.handle(writer, request)
		default:
			writer.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/delete", func(writer http.ResponseWriter, request *http.Request) {
		switch request.Method {
		case http.MethodPost:
			w.handleDelete(writer, request)
		default:
			writer.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/reset-password", func(writer http.ResponseWriter, request *http.Request) {
		switch request.Method {
		case http.MethodPost:
			w.handleResetPassword(writer, request)
		default:
			writer.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
	return mux
}

type Params struct {
	Location string
	User     string
	Password string
	Year     int64
	Score    float64
	Csrf     string
}

func (params *Params) Parse(request *http.Request) error {
	params.Location = request.FormValue("location")
	params.User = request.FormValue("user")
	params.Password = request.FormValue("password")
	if v, err := strconv.ParseInt(request.FormValue("year"), 10, 64); err != nil {
		return fmt.Errorf("parse year: %w", err)
	} else {
		params.Year = v
	}
	if v, err := strconv.ParseFloat(request.FormValue("score"), 64); err != nil {
		return fmt.Errorf("parse score: %w", err)
	} else {
		params.Score = v
	}
	params.Csrf = request.FormValue("csrf")
	return nil
}

type DeleteParams struct {
	User  string
	User2 string
}

func (params *DeleteParams) Parse(request *http.Request) error {
	params.User = request.FormValue("user")
	params.User2 = request.FormValue("user2")
	return nil
}

type ResetPasswordParams struct {
	Password string
	User     string
}

func (params *ResetPasswordParams) Parse(request *http.Request) error {
	params.Password = request.FormValue("password")
	params.User = request.FormValue("user")
	return nil
}

type wrappedResponse struct {
	headersSent bool
	real        http.ResponseWriter
}

func (wr *wrappedResponse) Header() http.Header {
	return wr.real.Header()
}

func (wr *wrappedResponse) Write(bytes []byte) (int, error) {
	wr.headersSent = true
	return wr.real.Write(bytes)
}

func (wr *wrappedResponse) WriteHeader(statusCode int) {
	wr.headersSent = true
	wr.real.WriteHeader(statusCode)
}

type wrapper struct {
	template   *template.Template
	controller Controller
}

func (wrp *wrapper) mainView(writer http.ResponseWriter, request *http.Request, lastError error) {
	data, err := wrp.controller.ViewData(request, lastError)
	if err != nil {
		log.Println("failed get view data:", err)
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}
	writer.Header().Set("Content-Type", "text/html")
	err = wrp.template.Execute(writer, data)
	if err != nil {
		log.Println("failed to render:", err)
	}
}
func (wrp *wrapper) handle(writer http.ResponseWriter, request *http.Request) {
	var params Params
	err := params.Parse(request)
	if err != nil {
		log.Println("failed paras  params:", err)
		writer.WriteHeader(http.StatusBadRequest)
		wrp.mainView(writer, request, err)
		return
	}

	wr := &wrappedResponse{real: writer}
	err = wrp.controller.Do(wr, request, params)
	if err != nil {
		log.Println("failed process :", err)
		writer.WriteHeader(http.StatusUnprocessableEntity)
		wrp.mainView(writer, request, err)
		return
	}
	if !wr.headersSent {
		writer.Header().Set("Location", ".")
		writer.WriteHeader(http.StatusSeeOther)
	}
}
func (wrp *wrapper) handleDelete(writer http.ResponseWriter, request *http.Request) {
	var params DeleteParams
	err := params.Parse(request)
	if err != nil {
		log.Println("failed paras Delete params:", err)
		writer.WriteHeader(http.StatusBadRequest)
		wrp.mainView(writer, request, err)
		return
	}

	wr := &wrappedResponse{real: writer}
	err = wrp.controller.DoDelete(wr, request, params)
	if err != nil {
		log.Println("failed process Delete:", err)
		writer.WriteHeader(http.StatusUnprocessableEntity)
		wrp.mainView(writer, request, err)
		return
	}
	if !wr.headersSent {
		writer.Header().Set("Location", ".")
		writer.WriteHeader(http.StatusSeeOther)
	}
}
func (wrp *wrapper) handleResetPassword(writer http.ResponseWriter, request *http.Request) {
	var params ResetPasswordParams
	err := params.Parse(request)
	if err != nil {
		log.Println("failed paras ResetPassword params:", err)
		writer.WriteHeader(http.StatusBadRequest)
		wrp.mainView(writer, request, err)
		return
	}

	wr := &wrappedResponse{real: writer}
	err = wrp.controller.DoResetPassword(wr, request, params)
	if err != nil {
		log.Println("failed process ResetPassword:", err)
		writer.WriteHeader(http.StatusUnprocessableEntity)
		wrp.mainView(writer, request, err)
		return
	}
	if !wr.headersSent {
		writer.Header().Set("Location", ".")
		writer.WriteHeader(http.StatusSeeOther)
	}
}
