package main

import (
	"net/http"
)

func handleHealth(writer http.ResponseWriter, req *http.Request) {
	writer.WriteHeader(200)

	_, err := writer.Write([]byte("OK"))
	if err != nil {
		return
	}

}
