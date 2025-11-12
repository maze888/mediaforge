#pragma once

#include <curl/curl.h>

int file_download(const char *url, const char *download_path);
