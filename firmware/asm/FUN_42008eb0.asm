; OTA info parser — 3_le_light_zb1_pid_55_v2.3.18.elf
; Source: Ghidra disassemble_function @ 0x42008eb0, strings verified by lepro.tools.validate_ota_parser
;
; Register args at entry:
;   a2 = cJSON* root object
;   a3 = secondary mode (stored ctx+0x19 on success)
;   a4 = OTA context struct
; Return: a2 = 0 success, 3 failure

42008eb0: entry a1,0x20
42008eb3: bnez.n a2,0x42008eb9
42008eb5: j 0x42008f60                    ; null object -> return 3

; --- fwType (required string, max 16) ---
42008eb9: l32r a11,0x420009ac             ; -> "fwType"
42008ebc: mov a10,a2
42008ebf: call8 0x42038c8c                ; cJSON_GetObjectItem(obj, "fwType")
42008ec5: bnez a10,0x42008ed5
42008ec8: call8 0x42009728
42008ecd: bnez.n a10,0x42008ef5           ; alternate success path
42008ecf: j 0x42008f60

42008ed5: l32r a8,0x420009d4              ; -> cJSON_IsString @ 0x420b855c
42008ed8: callx8 a8
42008edb: beqz a10,0x42008ec8
42008ede: l32i.n a10,a4,0x10
42008ee0: movi a12,0x10                   ; max length 16
42008ee3: movi a11,0x0
42008ee6: l32r a8,0x420009d8              ; -> validate_string_len @ 0x400014f4
42008ee9: callx8 a8
42008eef: beqz a10,0x42008ec8

; --- do_check (optional number, default 1) ---
42008ef5: l32r a11,0x420009b0             ; -> "do_check"
42008efb: call8 0x42038c8c
42008f01: bnez a10,0x42008f09
42008f04: movi.n a5,0x1                   ; a5 = counter baseline
42008f09: l32r a8,0x420009dc              ; -> cJSON_IsNumber @ 0x420b8548
42008f0c: callx8 a8
42008f0f: beqz a10,0x42008f04             ; not a number -> use default 1
42008f12: l8ui a5,a5,0x14                 ; cJSON.number.valueint

; --- version (required string; does NOT increment counter) ---
42008f15: l32r a11,0x420009b4             ; -> "version"
42008f18: s8i a5,a4,0x1b                  ; ctx->do_check
42008f1d: call8 0x42038c8c
42008f22: beqz.n a10,0x42008f65           ; missing -> counter=0, fail later
42008f24: l32r a8,0x420009d4
42008f27: callx8 a8
42008f2a: beqz.n a10,0x42008f65
42008f36: l32i.n a10,a5,0x10
42008f38: l32r a8,0x420009e0              ; -> cJSON_GetStringValue @ 0x40001380
42008f3b: callx8 a8
42008f41: s32i.n a10,a4,0x24              ; ctx->version

; --- hash (required string) ---
42008f6c: l32r a11,0x420009b8             ; -> "hash"
42008f71: call8 0x42038c8c
42008f76: beqz.n a10,0x42008f96
42008f78: l32r a8,0x420009d4
42008f7b: callx8 a8
42008f7e: beqz.n a10,0x42008f96
42008f8c: addi.n a5,a5,0x1                  ; counter++
42008f94: s32i.n a10,a4,0x20              ; ctx->hash

; --- path (required string) ---
42008f96: l32r a11,0x420009bc             ; -> "path"
42008f9b: call8 0x42038c8c
42008fa0: beqz.n a10,0x42008fc5
42008fa2: l32r a8,0x420009d4
42008fa5: callx8 a8
42008fa8: beqz.n a10,0x42008fc5
42008fb9: addi a5,a5,0x1                    ; counter++
42008fc2: s32i a10,a4,0x28                ; ctx->path

; --- size (required as STRING, max 10 chars) ---
42008fc5: l32r a11,0x420009c0             ; -> "size"
42008fca: call8 0x42038c8c
42008fd4: l32r a8,0x420009d4              ; cJSON_IsString — JSON number fails here
42008fd7: callx8 a8
42008fe2: l32i.n a10,a6,0x10
42008fe4: movi.n a12,0xa                  ; max length 10
42008fe8: l32r a8,0x420009d8
42008feb: callx8 a8
42008fee: s32i.n a10,a4,0x1c              ; ctx->size
42008ff0: addi.n a5,a5,0x1                  ; counter++

; --- secret (required string, may be "") ---
42008ff9: l32r a11,0x420009c4             ; -> "secret"
42008fff: call8 0x42038c8c
42009005: bnez a10,0x42009025
42009008: l32r a8,0x420008ac              ; log helper
4200900e: l32r a11,0x420009c8             ; -> "le_ota"
42009011: l32r a12,0x420009cc             ; -> "E %u %s ota info error"
42009020: movi.n a2,0x3
42009022: j 0x42009052

42009025: l32r a8,0x420009d4
42009028: callx8 a8
4200902b: beqz a10,0x42009008
42009042: call8 0x4200cf38                ; store secret string

; --- success gate: a5 must equal 4 ---
42009045: bnei a5,0x4,0x42009008           ; 1(do_check)+hash+path+size
42009048: movi.n a2,0x1
4200904a: s8i a2,a4,0x1a
4200904d: s8i a3,a4,0x19
42009050: movi.n a2,0x0                   ; return 0

42008f60: movi.n a2,0x3                   ; return 3
42009052: retw.n
