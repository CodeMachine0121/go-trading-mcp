# Contract Traceability Matrix — 助理認得奇異筆、但不畫奇異筆

Contract: PRD.md
Design map: ARCH.md
Implementation: `cmd/server/tool_catalog_chart_drawing.go`、`README.md`
Oracle: Acceptance Criteria（7 clauses：5 AC＋2 BR）

> 靜態一致性稽核：以 PRD 的驗收條件為準，判讀測試斷言與程式路徑，不以整套測試綠燈為依據，也不執行自創情境。
> 「拒絕原話帶回」是外掛既有的單一路徑（`internal/application/tests/api_tool_application_test.go:61` `TestARefusalComesBackInTheTradingServicesOwnWords`），本切片只確認奇異筆走同一條轉送路。

## Clauses

T＝`cmd/server/tool_catalog_chart_drawing_test.go`，C＝`cmd/server/tool_catalog_chart_drawing.go`

| ID | Clause | Spec-expected (oracle) | Impl | Test | Test audit | Code audit | Jev test | Jev code | Status | Route |
|----|--------|------------------------|------|------|------------|------------|----------|----------|--------|-------|
| AC-1 | 看繪圖板的說明介紹奇異筆 | 說明寫明奇異筆的錨點是沿筆跡的一連串點 | C:12-14、C:34 | T:180 `Teach…`（「kind 是 brush」「一連串點」） | asserts-oracle | produces-oracle | asserts-oracle 0.79 | produces-oracle 0.76 | ✅ conforms | recheck (mid confidence) |
| AC-2 | 奇異筆的點時間不一定是 K 線起始時間 | 說明寫明點時間不一定是 K 線起始時間 | C:13 | T:180（「openTime 不一定是 K 線的起始時間」） | asserts-oracle | produces-oracle | asserts-oracle 0.93 | produces-oracle 0.65 | ✅ conforms | recheck (mid confidence) |
| AC-3 | 可畫的種類照舊 | 八種、各自點數不變 | C:64-66 | T:222 `SpellsOutEveryAllowedValue` | asserts-oracle | produces-oracle | asserts-oracle 0.60 | produces-oracle 0.92 | ✅ conforms | recheck (mid confidence) |
| AC-4 | 說明寫明不能畫奇異筆 | 只能使用者手畫、助理不能新增，建議改用可畫種類或請使用者自己畫 | C:51-52、C:67 | T:180、T:222 | asserts-oracle | produces-oracle | asserts-oracle 1.00 | produces-oracle 0.93 | ✅ conforms | accept |
| AC-5 | 仍送了奇異筆時照原話轉達拒絕 | 含奇異筆的一批原樣送出；拒絕原話帶回 | C:38-72（轉送）＋既有轉達路徑 | T:103 `a brush the trading service will refuse`＋`api_tool_application_test.go:61` | asserts-oracle | produces-oracle | shallow 0.36 | produces-oracle 0.21 | 🟠 mis-asserted（Jev） | escalate (disagree) |
| BR-1 | 外掛不檢查奇異筆規則 | 含奇異筆的繪圖不在外掛被擋，原樣交給交易服務 | C:38-72 | T:103 | asserts-oracle | produces-oracle | asserts-oracle 0.85 | unclear 0.17 | ❔ unclear（Jev） | escalate (disagree) |
| BR-2 | 能力清單不變 | 仍只有看、新增、移除三件 | C:26-95 | `tool_catalog_test.go:155-168`、`:161` | asserts-oracle | produces-oracle | asserts-oracle 0.89 | produces-oracle 1.00 | ✅ conforms | accept |

**Notes**
- AC-5：Jev 判 shallow（低信心）是因為轉達拒絕的測試是通用的、不是奇異筆專屬。外掛不解析 `drawings`，奇異筆與任何其他被拒的繪圖走同一條轉達路，專屬測試不會多驗到東西；保留現狀，交由 James 判斷。
- BR-1：Jev 的程式稽核只拿到說明檔，轉送邏輯在 `internal/domain/models/domains/api_tool_domain.go`（原樣放進 body，不看內容）；T:103 斷言 body 與輸入 JSON 相等、無錯誤，已釘住「不擋」。

## Orphans

| Code | Behavior | Note |
|------|----------|------|
| — | — | 無。README 新增段落只描述本切片行為 |

## Summary

✅ 5 conforms · 🔴 0 violations · 🟠 1 mis-asserted（Jev，低信心）· 🟡 0 partial · ❌ 0 gaps · ❔ 1 unclear（Jev）· ⚠️ 0 orphans
Conformance: 71%（自評 shadow 為 7/7）
Needs James: AC-5（disagree）、BR-1（disagree）
