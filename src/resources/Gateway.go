package resources

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

type Gateway struct {
	Port    int
	Handler *GatewayHandler
	Server  *http.Server
}

func (g *Gateway) shutdown() error {
	fmt.Println("Shutting down gateway at port " + strconv.Itoa(g.Port))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	return g.Server.Shutdown(ctx)
}
