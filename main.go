package main

import (
	"net/http"
)

func main() {
	router := http.NewServeMux()
	//router.HandleFunc("", handlerHealth)

	server := http.Server{}
	server.Addr = ":8080"
	server.Handler = router

	//handler := healthHandler{"OK"}

	router.Handle("/app/", http.StripPrefix("/app", http.FileServer(http.Dir("."))))
	//router.Handle("/home", http.FileServer(http.Dir(home)))

	router.HandleFunc("/healthz/", handleHealth)

	err := server.ListenAndServe()
	if err != nil {
		//fmt.Print(err)
		return
	}
}
