package main

import (
	"context"
	"fmt"
	"golang.org/x/sync/errgroup"
	"io"
	"net/http"
	"os"
	"os/signal"
)

var shutdownChan = make(chan struct{})

func main() {
	ctx := context.Background()
	ctx, cancelFunc := context.WithCancel(ctx)
	group, errCtx := errgroup.WithContext(ctx)

	svr := &http.Server{Addr: ":9966"}

	group.Go(func() error {
		return startServer(svr)
	})

	group.Go(func() error {
		select {
		case <-shutdownChan:
			fmt.Println("shut down by user")
		case <-errCtx.Done():
			fmt.Println("shut down internal")
		}
		return svr.Shutdown(errCtx)
	})

	chanel := make(chan os.Signal, 1)
	signal.Notify(chanel)

	group.Go(func() error {
		select {
		case <-errCtx.Done():
			return errCtx.Err()
		case <-chanel:
			cancelFunc()
		}
		return nil
	})

	if err := group.Wait(); err != nil {
		fmt.Println("err: ", err)
	}
	fmt.Println("end")
}

func startServer(svr *http.Server) error {
	http.HandleFunc("/week03", handlerWeek03)
	http.HandleFunc("/shutdown", handlerShutDown)
	return svr.ListenAndServe()
}

func handlerShutDown(writer http.ResponseWriter, request *http.Request) {
	shutdownChan <- struct{}{}
}

func handlerWeek03(writer http.ResponseWriter, request *http.Request) {
	io.WriteString(writer, "hello week03.")
}
