#include "main.h"
#define DLL_EXPORT
extern "C" {
    int Runme(int var) {
        return var;
    }
}