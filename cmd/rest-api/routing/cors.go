package routing

import "net/http"

func EnableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	(*w).Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
}

func OptionsHandler(w http.ResponseWriter, r *http.Request) {
	EnableCors(&w)
	w.WriteHeader(http.StatusOK)
}
