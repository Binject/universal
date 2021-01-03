#ifndef MAIN_DLL_H
#define MAIN_DLL_H
#if defined DLL_EXPORT
#define BUILDING_MAIN_DLL __declspec(dllexport)
#else
#define BUILDING_MAIN_DLL __declspec(dllimport)
#endif
extern "C" {
    int Runme(int var);
}
#endif