#include "textflag.h"

// VRH:
//   k  = SplitMix64(key)          // calculated in Go
//   x  = node ^ k
//   x += ROTL(k, 23)
//   x ^= x >> 29
//   x += ROTL(k, 41)
//   x ^= x << 17
//   x ^= x >> 32
//
// Register aliasing note:
//
// Xn and Yn refer to the SAME physical register (Xn is simply the low
// 128 bits of Yn). A VEX-encoded write to Xn (e.g. VMOVQ reg, Xn) zeroes
// the upper 128 bits of the corresponding Yn as part of the AVX
// "zero-upper" rule. That means Xn must never be reused as a scratch
// register for building a DIFFERENT Yn once Yn already holds a live
// broadcast value: doing so silently destroys it. Each broadcast below
// therefore uses the scratch Xn whose index matches its own target Yn
// (X1->Y1, X2->Y2, X3->Y3) so no register is ever clobbered after use.

// func scoreVRHAVX2(k uint64, nodes *uint64, out *uint64, n int)
TEXT ·scoreVRHAVX2(SB), NOSPLIT, $0-32
	MOVQ k+0(FP), AX
	MOVQ nodes+8(FP), SI
	MOVQ out+16(FP), DI
	MOVQ n+24(FP), CX

	CMPQ CX, $0
	JE score_done

	// R8 = ROTL(k, 23)
	MOVQ AX, R8
	SHLQ $23, R8
	MOVQ AX, R9
	SHRQ $41, R9
	ORQ R9, R8

	// R10 = ROTL(k, 41)
	MOVQ AX, R10
	SHLQ $41, R10
	MOVQ AX, R9
	SHRQ $23, R9
	ORQ R9, R10

	MOVQ R8, X1
	VPBROADCASTQ X1, Y1

	MOVQ R10, X2
	VPBROADCASTQ X2, Y2

	MOVQ AX, X3
	VPBROADCASTQ X3, Y3

score_vector_loop:
	CMPQ CX, $4
	JB score_scalar_tail

	VMOVDQU (SI), Y0

	// x = node ^ k
	VPXOR Y3, Y0, Y0

	// x += k23
	VPADDQ Y1, Y0, Y0

	// x ^= x >> 29
	VPSRLQ $29, Y0, Y4
	VPXOR Y4, Y0, Y0

	// x += k41
	VPADDQ Y2, Y0, Y0

	// x ^= x << 17
	VPSLLQ $17, Y0, Y4
	VPXOR Y4, Y0, Y0

	// x ^= x >> 32
	VPSRLQ $32, Y0, Y4
	VPXOR Y4, Y0, Y0

	VMOVDQU Y0, (DI)

	ADDQ $32, SI
	ADDQ $32, DI
	SUBQ $4, CX
	JMP score_vector_loop

score_scalar_tail:
	TESTQ CX, CX
	JZ score_done

score_tail_loop:
	MOVQ (SI), DX

	// x = node ^ k
	XORQ AX, DX

	// x += k23
	ADDQ R8, DX

	// x ^= x >> 29
	MOVQ DX, R9
	SHRQ $29, R9
	XORQ R9, DX

	// x += k41
	ADDQ R10, DX

	// x ^= x << 17
	MOVQ DX, R9
	SHLQ $17, R9
	XORQ R9, DX

	// x ^= x >> 32
	MOVQ DX, R9
	SHRQ $32, R9
	XORQ R9, DX

	MOVQ DX, (DI)

	ADDQ $8, SI
	ADDQ $8, DI
	DECQ CX
	JNZ score_tail_loop

score_done:
	VZEROUPPER
	RET

// func ownerVRHAVX2(k uint64, nodes *uint64, n int) uint64
TEXT ·ownerVRHAVX2(SB), NOSPLIT, $64-32
	MOVQ k+0(FP), AX
	MOVQ nodes+8(FP), SI
	MOVQ n+16(FP), CX

	CMPQ CX, $0
	JE owner_empty

	MOVQ (SI), DX
	XORQ AX, DX

	// k23
	MOVQ AX, R8
	SHLQ $23, R8
	MOVQ AX, R9
	SHRQ $41, R9
	ORQ R9, R8

	ADDQ R8, DX

	MOVQ DX, R9
	SHRQ $29, R9
	XORQ R9, DX

	// k41
	MOVQ AX, R10
	SHLQ $41, R10
	MOVQ AX, R9
	SHRQ $23, R9
	ORQ R9, R10

	ADDQ R10, DX

	MOVQ DX, R9
	SHLQ $17, R9
	XORQ R9, DX

	MOVQ DX, R9
	SHRQ $32, R9
	XORQ R9, DX

	MOVQ DX, R12      // best score
	MOVQ (SI), R11    // best node

	CMPQ CX, $1
	JBE owner_done

	ADDQ $8, SI
	SUBQ $1, CX

	// Broadcast constants. Each Xn is only ever used to build its own
	// Yn (see note above scoreVRHAVX2) so Y1/Y2/Y3 all survive intact.
	MOVQ R8, X1
	VPBROADCASTQ X1, Y1

	MOVQ R10, X2
	VPBROADCASTQ X2, Y2

	MOVQ AX, X3
	VPBROADCASTQ X3, Y3

owner_vector_loop:
	CMPQ CX, $4
	JB owner_scalar_tail

	VMOVDQU (SI), Y0

	// x = node ^ k
	VPXOR Y3, Y0, Y0

	// x += k23
	VPADDQ Y1, Y0, Y0

	// x ^= x >> 29
	VPSRLQ $29, Y0, Y4
	VPXOR Y4, Y0, Y0

	// x += k41
	VPADDQ Y2, Y0, Y0

	// x ^= x << 17
	VPSLLQ $17, Y0, Y4
	VPXOR Y4, Y0, Y0

	// x ^= x >> 32
	VPSRLQ $32, Y0, Y4
	VPXOR Y4, Y0, Y0

	// Save scores to stack for scalar reduction.
	VMOVDQU Y0, 32(SP)

	// Save nodes.
	VMOVDQU (SI), Y5
	VMOVDQU Y5, 0(SP)

	// lane 0
	MOVQ 32(SP), R9
	CMPQ R9, R12
	JBE owner_lane1
	MOVQ R9, R12
	MOVQ 0(SP), R11

owner_lane1:
	MOVQ 40(SP), R9
	CMPQ R9, R12
	JBE owner_lane2
	MOVQ R9, R12
	MOVQ 8(SP), R11

owner_lane2:
	MOVQ 48(SP), R9
	CMPQ R9, R12
	JBE owner_lane3
	MOVQ R9, R12
	MOVQ 16(SP), R11

owner_lane3:
	MOVQ 56(SP), R9
	CMPQ R9, R12
	JBE owner_next
	MOVQ R9, R12
	MOVQ 24(SP), R11

owner_next:
	ADDQ $32, SI
	SUBQ $4, CX
	JMP owner_vector_loop

owner_scalar_tail:
	TESTQ CX, CX
	JZ owner_done

owner_tail_loop:
	MOVQ (SI), DX
	MOVQ DX, R13

	// x = node ^ k
	XORQ AX, DX

	// x += k23
	ADDQ R8, DX

	// x ^= x >> 29
	MOVQ DX, R9
	SHRQ $29, R9
	XORQ R9, DX

	// x += k41
	ADDQ R10, DX

	// x ^= x << 17
	MOVQ DX, R9
	SHLQ $17, R9
	XORQ R9, DX

	// x ^= x >> 32
	MOVQ DX, R9
	SHRQ $32, R9
	XORQ R9, DX

	CMPQ DX, R12
	JBE owner_tail_next

	MOVQ DX, R12
	MOVQ R13, R11

owner_tail_next:
	ADDQ $8, SI
	DECQ CX
	JNZ owner_tail_loop

owner_done:
	MOVQ R11, ret+24(FP)
	VZEROUPPER
	RET

owner_empty:
	XORQ AX, AX
	MOVQ AX, ret+24(FP)
	VZEROUPPER
	RET
