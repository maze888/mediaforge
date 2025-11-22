package tests

import (
    "testing"
)

func TestConvertMP4(t *testing.T) {
    jobDatas, err := helperMakeJobData(t, "./test_source/mp4", "mp4", "mp3")     
    if err != nil {
        t.Fatalf("helperMakeJobData() failed: %v", err)
    }

    helperSendMultipart(t, "http://localhost:5000/convert", jobDatas)
}
