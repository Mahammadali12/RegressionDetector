package sink

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regressiondetector/internal/collector/types"
)

type HttpSink struct{
	token string
	url string
}

func NewHttpSink(url, token string) *HttpSink {
	return &HttpSink{url: url, token: token}
}

func(s HttpSink) Write(ctx context.Context, records []types.PgStatRow) error{


	data, err := json.Marshal(records)

	if err != nil {
	    return fmt.Errorf("failed to serialize records: %w", err)
	}

	var writer bytes.Buffer

	gz := gzip.NewWriter(&writer)
	
	if _, err := gz.Write(data); err!=nil{
		// log.Fatal(err)
		return fmt.Errorf("failed to compress payload: %w", err)
	}
	if err := gz.Close(); err != nil { // Important: Close flushes buffered data
		// log.Fatal(err)
		return fmt.Errorf("failed to close writer: %w", err)
	}

	request, err := http.NewRequestWithContext(ctx,"POST",s.url,&writer)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)		
	}

	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s",s.token))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Content-Encoding", "gzip")

	resp, err := http.DefaultClient.Do(request)

	if err != nil {
    	return fmt.Errorf("failed to send payload: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
	    return fmt.Errorf("backend returned unexpected status: %d", resp.StatusCode)
	}

	return  nil

}