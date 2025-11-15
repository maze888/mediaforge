#pragma once

#include <curl/curl.h>

int http_file_download(const char *url, const char *file_path);
int http_file_upload(const char *url, const char *file_path);
