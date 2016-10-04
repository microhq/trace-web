package main

import (
	"github.com/micro/go-micro/client"
	"github.com/micro/go-web"
	trace "github.com/micro/trace-srv/proto/trace"
	"github.com/micro/trace-web/handler"
)

func main() {
	service := web.NewService(
		web.Name("go.micro.web.trace"),
		web.Handler(handler.Router()),
	)

	service.Init()

	handler.Init(
		"templates",
		trace.NewTraceClient("go.micro.srv.trace", client.DefaultClient),
	)

	service.Run()
}
