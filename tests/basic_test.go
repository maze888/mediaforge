package tests

import (
    "fmt"
    "testing"
    "encoding/json"
)

func TestConvertMP4(t *testing.T) {
    jobDatas, err := helperMakeJobData(t, "./test_source/mp4", "mp4", "mp3")     
    if err != nil {
        t.Fatalf("helperMakeJobData() failed: %v", err)
    }

    resp := helperSendMultipart(t, "http://localhost:5000/convert", jobDatas)
    defer resp.Body.Close()
    fmt.Printf("BODY: %+v\n", resp.Body)

    var jobResponses JobResponses
    if err := json.NewDecoder(resp.Body).Decode(&jobResponses); err != nil {
        t.Fatalf("json.NewDecoder().Decode() failed: %v", err)
    }

    for _, response := range jobResponses.Responses {
        fmt.Printf("jobID: %v\n", response.JobID)
        fmt.Printf("DownloadURL: %v\n", response.DownloadURL)
        downloadPresignedURL(response.DownloadURL)
    }
}
