#pragma once

#include <stdio.h>

#include <yyjson.h>

struct convert_request {
    char file_name[512];
    char input_format[32];
    char output_format[32];
    char upload_path[1024];
    char download_path[1024];
};

int json_parse(char *json, size_t json_len, struct convert_request *req); 
