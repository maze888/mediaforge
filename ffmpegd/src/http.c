#include <string.h>
#include <errno.h>

#include "http.h"

static size_t write_data(void *p, size_t size, size_t nmemb, void *stream) {
    return fwrite(p, size, nmemb, (FILE *)stream);
}

int http_file_download(const char *url, const char *file_path) {
    CURL *curl = NULL;
    FILE *fp = NULL;
    CURLcode res;

    curl = curl_easy_init();
    if (curl) {
        fp = fopen(file_path, "wb");
        if (!fp) {
            fprintf(stderr, "fopen() is failed: (path: %s errmsg: %s errno: %d)\n", file_path, strerror(errno), errno);
            goto out;
        }

        curl_easy_setopt(curl, CURLOPT_URL, url);
        curl_easy_setopt(curl, CURLOPT_WRITEFUNCTION, write_data);
        curl_easy_setopt(curl, CURLOPT_WRITEDATA, fp);

        // HTTPS 인증서 검증
        // curl_easy_setopt(curl, CURLOPT_SSL_VERIFYPEER, 1L);
        // curl_easy_setopt(curl, CURLOPT_SSL_VERIFYHOST, 2L);

        res = curl_easy_perform(curl);
        if (res != CURLE_OK) {
            fprintf(stderr, "curl_easy_perform() is failed: %s\n", curl_easy_strerror(res));
            goto out;
        }

        fclose(fp);
        curl_easy_cleanup(curl);
    }

    return 0;

out:
    if (fp) fclose(fp);
    if (curl) curl_easy_cleanup(curl);

    return -1;
}

int http_file_upload(const char *url, const char *file_path) {
    CURL *curl = NULL;
    FILE *fp = NULL;
    CURLcode res;

    curl = curl_easy_init();
    if (curl) {
        fp = fopen(file_path, "rb");
        if (!fp) {
            fprintf(stderr, "fopen() is failed: (path: %s errmsg: %s errno: %d)\n", file_path, strerror(errno), errno);
            goto out;
        }

        // 업로드할 presigned URL 지정
        curl_easy_setopt(curl, CURLOPT_URL, url);
        // 업로드 메서드 PUT 지정
        curl_easy_setopt(curl, CURLOPT_UPLOAD, 1L);
        // 업로드할 파일 지정
        curl_easy_setopt(curl, CURLOPT_READDATA, fp);

        // 파일 크기 설정(권장)
        fseek(fp, 0L, SEEK_END);
        long filesize = ftell(fp);
        rewind(fp);
        curl_easy_setopt(curl, CURLOPT_INFILESIZE_LARGE, (curl_off_t)filesize);
        
        // HTTPS 인증서 검증 (0L 임시 비활성화)
        curl_easy_setopt(curl, CURLOPT_SSL_VERIFYPEER, 0L);
        curl_easy_setopt(curl, CURLOPT_SSL_VERIFYHOST, 0L);

        // verbose 모드 - 디버깅용
        // curl_easy_setopt(curl, CURLOPT_VERBOSE, 1L);

        res = curl_easy_perform(curl);
        if (res != CURLE_OK) {
            fprintf(stderr, "curl_easy_perform() is failed: %s\n", curl_easy_strerror(res));
            goto out;
        }

        fclose(fp);
        curl_easy_cleanup(curl);
    }

    return 0;

out:
    if (fp) fclose(fp);
    if (curl) curl_easy_cleanup(curl);

    return -1;
}
