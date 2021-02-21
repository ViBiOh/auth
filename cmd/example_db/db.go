package main

import (
	"bytes"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/ViBiOh/auth/v2/pkg/ident/basic"
	"github.com/ViBiOh/auth/v2/pkg/middleware"
	dbStore "github.com/ViBiOh/auth/v2/pkg/store/db"
	"github.com/ViBiOh/httputils/v4/pkg/db"
	"github.com/ViBiOh/httputils/v4/pkg/health"
	"github.com/ViBiOh/httputils/v4/pkg/httperror"
	"github.com/ViBiOh/httputils/v4/pkg/httputils"
	"github.com/ViBiOh/httputils/v4/pkg/logger"
	"github.com/ViBiOh/httputils/v4/pkg/request"
	"github.com/ViBiOh/httputils/v4/pkg/server"
)

func main() {
	fs := flag.NewFlagSet("example", flag.ExitOnError)

	appServerConfig := server.Flags(fs, "")
	healthConfig := health.Flags(fs, "")

	dbConfig := db.Flags(fs, "db")

	logger.Fatal(fs.Parse(os.Args[1:]))

	appServer := server.New(appServerConfig)

	appDB, err := db.New(dbConfig)
	logger.Fatal(err)

	healthApp := health.New(healthConfig, appDB.Ping)

	authProvider := dbStore.New(appDB)
	identProvider := basic.New(authProvider)
	middlewareApp := middleware.New(authProvider, identProvider)

	go appServer.Start("http", healthApp.End(), httputils.Handler(Handler(), healthApp, middlewareApp.Middleware))

	healthApp.WaitForTermination(appServer.Done())
	server.GracefulWait(appServer.Done())
}

// Handler for dump request. Should be use with net/http
func Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		value := dumpRequest(r)

		logger.Info("Dump of request\n%s", value)

		if _, err := w.Write([]byte(value)); err != nil {
			httperror.InternalServerError(w, err)
		}
	})
}

func dumpRequest(r *http.Request) string {
	var headers bytes.Buffer
	for key, value := range r.Header {
		headers.WriteString(fmt.Sprintf("%s: %s\n", key, strings.Join(value, ",")))
	}

	var params bytes.Buffer
	for key, value := range r.URL.Query() {
		headers.WriteString(fmt.Sprintf("%s: %s\n", key, strings.Join(value, ",")))
	}

	var form bytes.Buffer
	if err := r.ParseForm(); err != nil {
		form.WriteString(err.Error())
	} else {
		for key, value := range r.PostForm {
			form.WriteString(fmt.Sprintf("%s: %s\n", key, strings.Join(value, ",")))
		}
	}

	body, err := request.ReadBodyRequest(r)
	if err != nil {
		logger.Error("%s", err)
	}

	var outputPattern bytes.Buffer
	outputPattern.WriteString("%s %s\n")
	outputData := []interface{}{
		r.Method,
		r.URL.Path,
	}

	if headers.Len() != 0 {
		outputPattern.WriteString("Headers\n%s\n")
		outputData = append(outputData, headers.String())
	}

	if params.Len() != 0 {
		outputPattern.WriteString("Params\n%s\n")
		outputData = append(outputData, params.String())
	}

	if form.Len() != 0 {
		outputPattern.WriteString("Form\n%s\n")
		outputData = append(outputData, form.String())
	}

	if len(body) != 0 {
		outputPattern.WriteString("Body\n%s\n")
		outputData = append(outputData, body)
	}

	return fmt.Sprintf(outputPattern.String(), outputData...)
}
