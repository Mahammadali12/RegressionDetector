package api

import (
	"encoding/json"
	"net/http"
	"regressiondetector/store"
)


func HandleGetAnomalies(s* store.Store)  http.HandlerFunc{
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET"{
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		//call store.GetAll
		//set Content-Type to application/json
		//encode response as json
		anomalies, err := s.GetAll(r.Context())
		if err != nil {
			http.Error(w, "failed to get anomalies", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"anomalies": anomalies,
		})
	}
}