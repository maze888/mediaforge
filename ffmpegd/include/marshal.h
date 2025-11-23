#pragma once

#include <stdio.h>
#include <yyjson.h>

#define FILENAME_LEN 512
#define PATH_LEN 1024

struct convert_request {
    char job_id[64];
    char file_name[FILENAME_LEN];
    char input_format[32];
    char output_format[32];
    char upload_path[PATH_LEN];
    char download_path[PATH_LEN];
};

struct convert_response {
    char job_id[64];
    char errmsg[256];
    bool status;
};

int json_unmarshal(char *json, size_t json_len, struct convert_request *req);
yyjson_mut_doc * json_marshal(struct convert_response *res);
