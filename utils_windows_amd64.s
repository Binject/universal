// stolen from https://github.com/C-Sto/BananaPhone
//func getModuleLoadedOrder(i int) (start uintptr, size uintptr)
TEXT ·getModuleLoadedOrder(SB), $0-32
	//All operations push values into AX
	//PEB
	MOVQ 0x60(GS), AX
	//PEB->LDR
	MOVQ 0x18(AX),AX

	//LDR->InMemoryOrderModuleList
	MOVQ 0x20(AX),AX

	//loop things
	XORQ R10,R10
startloop:
	CMPQ R10,i+0(FP)
	JE endloop
	//Flink (get next element)
	MOVQ (AX),AX
	INCQ R10
	JMP startloop
endloop:
	//Flink - 0x10 -> _LDR_DATA_TABLE_ENTRY
	//_LDR_DATA_TABLE_ENTRY->DllBase (offset 0x30)
	MOVQ 0x20(AX),CX
	MOVQ CX, start+8(FP)
	
	MOVQ 0x30(AX),CX
	MOVQ CX, size+16(FP)
	MOVQ AX,CX
	ADDQ $0x38,CX
	MOVQ CX, modulepath+24(FP)
	//SYSCALL
	RET 

