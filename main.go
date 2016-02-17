package main

import (
	"github.com/gorilla/mux"
	"github.com/micro/go-micro/client"
	"github.com/micro/go-web"
	"github.com/micro/trace-srv/proto/trace"
	"github.com/micro/trace-web/handler"
)

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", handler.Index)
	r.HandleFunc("/search", handler.Search)
	r.HandleFunc("/latest", handler.Latest)
	r.HandleFunc("/trace/{id}", handler.Trace)

	service := web.NewService(
		web.Name("go.micro.web.trace"),
		web.Handler(r),
	)

	service.Init()
	handler.TraceClient = trace.NewTraceClient("go.micro.srv.trace", client.DefaultClient)
	service.Run()
}
