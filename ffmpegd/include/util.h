#pragma once

#include <uuid/uuid.h>

#define CKNUL(p) p ? "OK" : "NULL"
#define safe_free(p) if (p) { free(p); p = NULL; }

void generate_uuid(char *uuid);
