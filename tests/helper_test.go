package tests

import (
    "fmt"
    "os"
    "io"
    "bytes"
    "testing"
    "net/http"
    "io/fs"
    "mime"
    "mime/multipart"
    "path/filepath"

    "github.com/google/uuid"
)

type JobData struct {
    JobID        string
    InputFormat  string
    OutputFormat string
    FilePath     string // 로컬에서 읽을 실제 파일 경로
}

type JobResponse struct {
    JobID string `json:"job_id" binding:"required"`
    DownloadURL string `json:"download_url" binding:"required"`
}

type JobResponses struct {
    Responses []JobResponse `json:"results" binding:"required"`
}

func helperMakeJobData(t *testing.T, srcPath, inputFormat, outputFormat string) (jobDatas []JobData, err error) {
    t.Helper()

    err = filepath.WalkDir(srcPath, func(path string, d fs.DirEntry, err error) error {
        if err != nil {
            return err
        }

        // 디렉토리 건너뜀
        if d.IsDir() {
            return nil
        }

        jobData := JobData {
            JobID: fmt.Sprintf("JobID-%s", uuid.NewString()),
            InputFormat: inputFormat,
            OutputFormat: outputFormat,
            FilePath: path,
        }
        jobDatas = append(jobDatas, jobData)

        return nil
    })
    if err != nil {
        t.Fatalf("filepath.WalkDir() failed: %v", err)
    }

    return jobDatas, nil
}

func helperSendMultipart(t *testing.T, url string, jobs []JobData) *http.Response {
    t.Helper()

    var requestBody bytes.Buffer

    // 요청 본문 생성 
    writer := multipart.NewWriter(&requestBody)
    for _, job := range jobs {
        // 1. JobID
        if err := writer.WriteField("JobID", job.JobID); err != nil {
            t.Fatalf("writer.WriteField(JobID): %v", err)
        }
        // 2. InputFormat
        if err := writer.WriteField("InputFormat", job.InputFormat); err != nil {
            t.Fatalf("writer.WriteField(InputFormat): %v", err)
        }
        // 3. OutputFormat
        if err := writer.WriteField("OutputFormat", job.OutputFormat); err != nil {
            t.Fatalf("writer.WriteField(OutputFormat): %v", err)
        }
        // 4. files
        if err := addFileToWriter(writer, "files", job.FilePath); err != nil {
            t.Fatalf("addFileToWriter() failed: %v", err)
        }
    }
    // 바운더리 닫고 요청 본문 완성
    if err := writer.Close(); err != nil {
        t.Fatalf("writer.Close() failed: %v", err)
    }

    // HTTP 요청 생성
    req, err := http.NewRequest("POST", url, &requestBody)
    if err != nil {
        t.Fatalf("http.NewRequest() failed: %v", err)
    }
    req.Header.Set("Content-Type", writer.FormDataContentType())

    // HTTP 요청 전송
    client := &http.Client {}
    resp, err := client.Do(req)
    if err != nil {
        t.Fatalf("client.Do() failed: %v", err)
    }

    return resp
}

func addFileToWriter(writer *multipart.Writer, fieldName, filePath string) error {
    file, err := os.Open(filePath)
    if err != nil {
        return fmt.Errorf("os.Open() failed: %v", err)
    }
    defer file.Close()

    part, err := writer.CreateFormFile(fieldName, filepath.Base(filePath))
    if err != nil {
        return fmt.Errorf("writer.CreateFormFile() failed: %v", err)
    }

    if _, err := io.Copy(part, file); err != nil {
        return fmt.Errorf("io.Copy() failed: %v", err)
    }

    return nil
}

func downloadPresignedURL(url string) error {
    resp, err := http.Get(url)
    if err != nil {
        return fmt.Errorf("http.Get() failed: %v", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("Failed to download file. Status: %v", resp.Status)
    }

    filename := ""
    if cd := resp.Header.Get("Content-Disposition"); cd != "" {
        if _, params, err := mime.ParseMediaType(cd); err == nil {
            if fn, ok := params["filename"]; ok {
                filename = fn
            }
        }
    }
    
    file, err := os.Create(filename)
    if err != nil {
        return fmt.Errorf("os.Create() failed: (filename: %v, err: %v)", filename, err)
    }
    defer file.Close()

    _, err = io.Copy(file, resp.Body)
    if err != nil {
        return fmt.Errorf("io.Copy() failed: %v", err)
    }

    return nil
}
