package main

import (
	"flag"
	"fmt"

	"fishing-notes-admin-api/internal/config"
	"fishing-notes-admin-api/internal/handler"
	"fishing-notes-admin-api/internal/responsex"
	"fishing-notes-admin-api/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/rest/httpx"
)

var configFile = flag.String("f", "etc/config.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	httpx.SetErrorHandlerCtx(responsex.ErrorHandler)
	httpx.SetOkHandler(responsex.OkHandler)

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	ctx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
