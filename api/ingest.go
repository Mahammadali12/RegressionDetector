package api

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regressiondetector/engine"
	"regressiondetector/internal/collector/types"
	"regressiondetector/store"
)

func HandleIngest(token string, store* store.Store, detector* engine.Detector) http.HandlerFunc{
	return func(w http.ResponseWriter, r *http.Request) {

		if r.Method != "POST"{
    		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
    		return
		}
		if r.Header.Get("Authorization") != fmt.Sprintf("Bearer %s",token){
			http.Error(w, "Forbidden", http.StatusUnauthorized)
    		return
		}
		if r.Header.Get("Content-Encoding") != "gzip"{
			http.Error(w, "wrong encoding", http.StatusNotAcceptable)
    		return
		}
		if r.Header.Get("Content-Type") != "application/json"{
			http.Error(w, "wrong file format", http.StatusNotAcceptable)
    		return
		}


		gz,err := gzip.NewReader(r.Body)
		if err != nil {
			http.Error(w, "failed to decompress payload", http.StatusBadRequest)
			return
		}


		data, err := io.ReadAll(gz);
		if  err != nil{
			http.Error(w, "failed to read payload", http.StatusBadRequest)
			return

		}

		if err := gz.Close(); err != nil { // Important: Close flushes buffered data
			http.Error(w, "failed to close reader", http.StatusInternalServerError)
			return 
		}

		var records []types.PgStatRow
		if err := json.Unmarshal(data, &records); err != nil {
		    http.Error(w, "failed to parse payload", http.StatusBadRequest)
		    return
		}	

		err = store.Save(r.Context(),records)
		if err != nil {
		    http.Error(w, "failed to save records", http.StatusInternalServerError)			
			return 
		}

		for _, record := range records {
			err := detector.Analyze(r.Context(), record)
    		if err != nil {
    		    log.Printf("analysis failed for query %d: %v", record.QueryID, err)
    		}
		}

	}
	
}