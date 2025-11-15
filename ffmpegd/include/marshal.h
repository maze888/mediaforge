#pragma once

#include <stdio.h>

#include <yyjson.h>

struct convert_request {
    char job_id[64];
    char file_name[512];
    char input_format[32];
    char output_format[32];
    char upload_path[1024];
    char download_path[1024];
};

struct convert_response {
    char job_id[64];
    char errmsg[256];
    bool ok;
};

int json_unmarshal(char *json, size_t json_len, struct convert_request *req);
yyjson_mut_doc * json_marshal(struct convert_response *res);
