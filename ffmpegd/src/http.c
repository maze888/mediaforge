#include <string.h>
#include <errno.h>

#include "http.h"

static size_t write_data(void *p, size_t size, size_t nmemb, void *stream) {
    return fwrite(p, size, nmemb, (FILE *)stream);
}

int file_download(const char *url, const char *download_path) {
    CURL *curl = NULL;
    FILE *fp = NULL;
    CURLcode res;

    curl = curl_easy_init();
    if (curl) {
        fp = fopen(download_path, "wb");
        if (!fp) {
            fprintf(stderr, "fopen() is failed: (errmsg: %s errno: %d)", strerror(errno), errno);
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
