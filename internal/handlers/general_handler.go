package handler

import "net/http"

type GeneralHandler struct {
	//------TODO
}

func NewGeneralHandler() *GeneralHandler {
	return &GeneralHandler{}
}

func (gh *GeneralHandler) HelloWorld(w http.ResponseWriter, r *http.Request) {
	content := "Hello World From UptimeMonitor"
	w.Write([]byte(content))
	w.WriteHeader(http.StatusOK)
}
