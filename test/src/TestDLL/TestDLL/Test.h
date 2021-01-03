// Test.h - Contains declarations of test functions
#pragma once

#ifdef TESTDLL_EXPORTS
#define TESTDLL_API __declspec(dllexport)
#else
#define TESTDLL_API __declspec(dllimport)
#endif

// Test function 
extern "C" TESTDLL_API int Runme(int a); 