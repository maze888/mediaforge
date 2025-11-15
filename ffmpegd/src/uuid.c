#include "util.h"

void generate_uuid(char *uuid) {
    if (uuid) {
        uuid_t uuid_raw;
        uuid_generate(uuid_raw);
        uuid_unparse_lower(uuid_raw, uuid);
    }
}
