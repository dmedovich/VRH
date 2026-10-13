#include "textflag.h"

// N3 transformed-node owner kernel, copied from the validated research prototype.
TEXT ·ownerN3Prepared(SB), NOSPLIT, $8-32
	MOVQ k+0(FP), AX
	MOVQ nodes+8(FP), SI
	MOVQ n+16(FP), CX
	MOVQ SI, 0(SP)
	XORQ R15, R15
	XORQ R12, R12
	MOVQ $0x2cb0f69f4abea221, R8
	MOVQ $0xb19e04cb9dce3147, R10
	MOVQ $0xbf58476d1ce4e5b9, BX
	CMPQ CX, $4
	JB pret_tail
pret_loop:
	MOVQ 0(SI), DX
	XORQ AX, DX
	MOVQ 8(SI), DI
	XORQ AX, DI
	MOVQ 16(SI), R11
	XORQ AX, R11
	MOVQ 24(SI), R13
	XORQ AX, R13
	IMULQ R8, DX
	IMULQ R8, DI
	IMULQ R8, R11
	IMULQ R8, R13
	MOVQ DX, R9
	SHRQ $27, R9
	XORQ R9, DX
	MOVQ DI, R9
	SHRQ $27, R9
	XORQ R9, DI
	MOVQ R11, R9
	SHRQ $27, R9
	XORQ R9, R11
	MOVQ R13, R9
	SHRQ $27, R9
	XORQ R9, R13
	ROLQ $23, DX
	ROLQ $23, DI
	ROLQ $23, R11
	ROLQ $23, R13
	IMULQ R10, DX
	IMULQ R10, DI
	IMULQ R10, R11
	IMULQ R10, R13
	MOVQ DX, R9
	SHRQ $31, R9
	XORQ R9, DX
	MOVQ DI, R9
	SHRQ $31, R9
	XORQ R9, DI
	MOVQ R11, R9
	SHRQ $31, R9
	XORQ R9, R11
	MOVQ R13, R9
	SHRQ $31, R9
	XORQ R9, R13
	IMULQ BX, DX
	IMULQ BX, DI
	IMULQ BX, R11
	IMULQ BX, R13
	CMPQ DX, R12
	JCC pret_record0
pret_next0:
	CMPQ DI, R12
	JCC pret_record1
pret_next1:
	CMPQ R11, R12
	JCC pret_record2
pret_next2:
	CMPQ R13, R12
	JCC pret_record3
pret_next3:
	ADDQ $32, SI
	SUBQ $4, CX
	CMPQ CX, $4
	JAE pret_loop
pret_tail:
	TESTQ CX, CX
	JZ pret_done
pret_tail_loop:
	MOVQ (SI), DX
	XORQ AX, DX
	IMULQ R8, DX
	MOVQ DX, R9
	SHRQ $27, R9
	XORQ R9, DX
	ROLQ $23, DX
	IMULQ R10, DX
	MOVQ DX, R9
	SHRQ $31, R9
	XORQ R9, DX
	IMULQ BX, DX
	MOVQ DX, R9
	SHRQ $33, R9
	XORQ R9, DX
	CMPQ DX, R15
	JLS pret_tail_next
	MOVQ DX, R15
	MOVQ SI, 0(SP)
pret_tail_next:
	ADDQ $8, SI
	DECQ CX
	JNZ pret_tail_loop
pret_done:
	MOVQ 0(SP), R9
	MOVQ (R9), AX
	MOVQ AX, ret+24(FP)
	RET
pret_record0:
	MOVQ DX, R9
	SHRQ $33, R9
	XORQ R9, DX
	CMPQ DX, R15
	JLS pret_next0
	MOVQ DX, R15
	LEAQ 0(SI), R9
	MOVQ R9, 0(SP)
	MOVQ R15, R12
	SHRQ $31, R12
	SHLQ $31, R12
	JMP pret_next0
pret_record1:
	MOVQ DI, R9
	SHRQ $33, R9
	XORQ R9, DI
	CMPQ DI, R15
	JLS pret_next1
	MOVQ DI, R15
	LEAQ 8(SI), R9
	MOVQ R9, 0(SP)
	MOVQ R15, R12
	SHRQ $31, R12
	SHLQ $31, R12
	JMP pret_next1
pret_record2:
	MOVQ R11, R9
	SHRQ $33, R9
	XORQ R9, R11
	CMPQ R11, R15
	JLS pret_next2
	MOVQ R11, R15
	LEAQ 16(SI), R9
	MOVQ R9, 0(SP)
	MOVQ R15, R12
	SHRQ $31, R12
	SHLQ $31, R12
	JMP pret_next2
pret_record3:
	MOVQ R13, R9
	SHRQ $33, R9
	XORQ R9, R13
	CMPQ R13, R15
	JLS pret_next3
	MOVQ R13, R15
	LEAQ 24(SI), R9
	MOVQ R9, 0(SP)
	MOVQ R15, R12
	SHRQ $31, R12
	SHLQ $31, R12
	JMP pret_next3
