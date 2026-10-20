# Contract Traceability Matrix — 助理在使用者的 K 線圖上畫線（外掛）

Contract: PRD.md
Design map: ARCH.md
Implementation: `cmd/server/tool_catalog_chart_drawing.go`、`cmd/server/tool_catalog.go`、`cmd/server/dependencies.go`、`README.md`
Oracle: Acceptance Criteria（19 clauses：16 AC＋3 BR）＋ 2 NFR

> 靜態一致性稽核：以 PRD 的驗收條件為準，判讀測試斷言與程式路徑，不以整套測試綠燈為依據，也不執行自創情境。
> 「拒絕原話帶回」「連不到交易服務」「重新連線」是外掛既有的單一路徑（`TradingServiceProxy.responseOf`，`internal/application/tests/api_tool_application_test.go:61` 起），本切片只確認繪圖能力走同一條路。
> Jev 第一輪的測試範圍只給到資料列、沒含斷言迴圈，AC-2／3／5／6／7／8／10／11／13 以修正後的完整範圍重跑；下表記的是重跑結果。

## Clauses

T＝`cmd/server/tool_catalog_chart_drawing_test.go`，C＝`cmd/server/tool_catalog_chart_drawing.go`，A＝`internal/domain/models/domains/api_tool_domain.go`

| ID | Clause | Spec-expected (oracle) | Impl | Test | Test audit | Code audit | Jev test | Jev code | Status | Route |
|----|--------|------------------------|------|------|------------|------------|----------|----------|--------|-------|
| AC-1 | 兩種來源都列出並標明來源 | 讀的是合約、btcusdt 那一份；帶回標明來源與版本 | C:24-33 | T:24 `AsksTheRightBoard`、T:177 `Teach…` | asserts-oracle | produces-oracle | asserts-oracle 0.73 | produces-oracle 0.43 | ✅ conforms | escalate (low confidence) |
| AC-2 | 從沒畫過的繪圖板是空的一份 | 空的一份、版本 0，不是錯誤 | C:28 | T:177 | asserts-oracle | produces-oracle | asserts-oracle 0.37 | not-implemented 0.20 | ❌ gap（Jev） | escalate (disagree) |
| AC-3 | 市場說錯由交易服務拒絕並原話帶回 | 照送、拒絕原話帶回；說明寫明只有現貨與合約 | C:15-16、A:142 | T:36（期貨跳脫後照送）、T:229 | asserts-oracle | produces-oracle | asserts-oracle 0.41 | diverges 0.33 | 🔴 violation（Jev） | escalate (violation) |
| AC-4 | 吸附最低價畫一條水平線 | 刻度與繪圖原樣交給合約 BTCUSDT 繪圖板；說明要求提醒回網頁看 | C:34-64、C:12 | T:24、T:88、T:177 | asserts-oracle | produces-oracle | asserts-oracle 0.24 | produces-oracle 0.44 | ✅ conforms | escalate (low confidence) |
| AC-5 | 一次二十個照常送出 | 20 個一個不少交給交易服務 | C:58-64 | T:88 `twenty drawings at once` | asserts-oracle | produces-oracle | asserts-oracle 0.84 | produces-oracle 0.63 | ✅ conforms | recheck (mid confidence) |
| AC-6 | 時間對不上刻度由交易服務整批拒絕並原話帶回 | 照送；說明要照拒絕改再送、不原樣重送 | C:47-48 | T:88、T:177 | asserts-oracle | produces-oracle | asserts-oracle 0.65 | unclear 0.05 | ❔ unclear（Jev） | escalate (disagree) |
| AC-7 | 空位不夠由交易服務整批拒絕並說出剩幾個 | 拒絕原話；說明寫明上限與剩幾個空位 | C:46 | T:177 | asserts-oracle | produces-oracle | asserts-oracle 0.81 | produces-oracle 0.47 | ✅ conforms | escalate (low confidence) |
| AC-8 | 沒說刻度外掛先擋下 | 不送出，回覆刻度是必填 | C:51-52、A:145 | T:120 | asserts-oracle | produces-oracle | asserts-oracle 0.36 | produces-oracle 0.40 | ✅ conforms | escalate (low confidence) |
| AC-9 | 說明教會助理畫對 | 說明含讀 K 線能力、對齊起始時間與六種刻度、吸附優先、八種與點數、樣式、上限、只新增助手繪圖、不必先問 | C:34-64 | T:177、T:209 | asserts-oracle | produces-oracle | asserts-oracle 0.88 | produces-oracle 0.95 | ✅ conforms | accept |
| AC-10 | 指名移除一條助手畫的 | 「只移除 B」交給交易服務 | C:74-76 | T:24（ids=12） | asserts-oracle | produces-oracle | asserts-oracle 0.69 | produces-oracle 0.76 | ✅ conforms | recheck (mid confidence) |
| AC-11 | 不指名即移除助手畫的全部 | 不帶任何編號 | C:74-76 | T:24（無 query） | asserts-oracle | produces-oracle | asserts-oracle 0.91 | produces-oracle 0.93 | ✅ conforms | accept |
| AC-12 | 指名到手畫的整次拒絕 | 拒絕原話、一個沒刪；說明叫助理請使用者網頁上動、不改用別的辦法 | C:8-10、C:69 | T:177 | asserts-oracle | produces-oracle | asserts-oracle 0.80 | produces-oracle 0.51 | ✅ conforms | escalate (low confidence) |
| AC-13 | 指名不存在的編號整次拒絕 | 拒絕原話帶回 | C:69、C:74 | T:24（ids 照送） | shallow | produces-oracle | asserts-oracle 0.96 | produces-oracle 0.75 | ✅ conforms（Jev） | escalate (disagree) |
| AC-14 | 能力清單只有看、畫、移除助手畫的三件 | 正好三件；沒有整份存回或改動繪圖板 | C:22-78、`tool_catalog.go` apiToolCatalog | T:156 `NoAbilityCanRewrite…`、`tool_catalog_test.go` 總表 | asserts-oracle | produces-oracle | asserts-oracle 0.51 | produces-oracle 0.95 | ✅ conforms | escalate (low confidence) |
| AC-15 | 使用者要求改手畫的線 | 說明告訴助理沒有修改能力、請使用者網頁上動 | C:8-10 | T:177 | asserts-oracle | produces-oracle | asserts-oracle 0.88 | produces-oracle 0.92 | ✅ conforms | accept |
| AC-16 | 開場說明與專案說明提到畫線 | 開場說明提到畫線與只動助手畫的；README 列三件 | `dependencies.go` Instructions、README〈在 K 線圖上畫線〉 | `dependencies_test.go` `TestTheConnectorInstructionsSayItDrawsOnlyAssistantLines` | asserts-oracle（README 無測試） | produces-oracle | not run | not run | 🟡 partial（README 部分） | — |
| BR-1 | 外掛只轉達：網址、內文、查詢字串原樣 | 市場與標的進網址、刻度與繪圖進內文、編號進查詢字串 | C:14-78、A:142-180 | T:24、T:68、T:88 | asserts-oracle | produces-oracle | not run | not run | ✅ conforms | — |
| BR-2 | 規則全由交易服務決定 | 外掛不檢查刻度、種類、樣式、上限 | C 只有宣告 | T:88（不合規的 08:30 照送） | asserts-oracle | produces-oracle | not run | not run | ✅ conforms | — |
| BR-3 | 外掛只擋必填 | 市場、標的、刻度、繪圖必填，ids 選填 | C:14-20、51-76 | T:120、T:147、T:229 | asserts-oracle | produces-oracle | not run | not run | ✅ conforms | — |
| NFR-1 | 以呼叫端外掛授權進行、沒有指定別人的欄位、無整份存回 | 欄位只有 market/symbol/interval/drawings/ids | C:14-78 | T:156、`tool_catalog_test.go` 總表 | asserts-oracle | produces-oracle | not run | not run | ✅ conforms | — |
| NFR-2 | 一般等待時間即可 | 不另設長等待 | C（無 Waiting） | `tool_catalog_test.go` 只有看一眼停在線上的守門 | shallow | produces-oracle | not run | not run | 🟠 mis-asserted（無直接斷言） | — |

**Claude 的判讀（Jev 低信心／不一致）**：
- AC-2：「revision 0、drawings 是空的」寫在 C:28，T:177 斷言 `revision 0`；Jev 判 not-implemented 信心 0.20，判為誤判。
- AC-3：路徑裡的「期貨」以網址跳脫照送（`%E6%9C%9F%E8%B2%A8`），那是傳輸正確，不是改寫；拒絕走既有原話路徑。Jev diverges 信心 0.33，判為誤判。
- AC-6：程式不檢查時間（照送），Jev unclear 信心 0.05。判為 conforms。
- AC-13：自評 shallow——沒有斷言說明寫明「沒有的編號也整次拒絕」。已補（見下方回饋）。

## Orphans (code with no clause)

| Code | Description | Verdict |
|------|-------------|---------|
| T:68 `ReadingOrRemovingSendsNothingInTheContents` | 看與移除不送內文 | undocumented（BR-1 的延伸，保留） |
| README 能力名稱不帶 assistant 的說明 | 解釋與既有守門的關係 | undocumented（ARCH 已記錄的決定） |

沒有任何程式碼落在 Out of Scope（整份存回、修改、外掛自己檢查、即時推播、新讀行情能力）。

## Summary

- Conforms（含 Claude 判讀覆核）: 17/21 ✅
- Violations: AC-3（Jev 低信心，Claude 判為誤判）
- Mis-asserted: NFR-2（沒有直接斷言不等久）
- Partial: AC-16（README 無測試，文件不測）
- Gaps: AC-2（Jev 低信心，Claude 判為誤判）
- Unclear: AC-6（Jev 低信心）
- Orphans: 2（皆非範圍外）
- Jev: ran on 15 clauses（AC-1～AC-15）
- Needs James（escalate，供使用者複核）: AC-1、2、3、4、6、7、8、12、13、14（escalate：低信心或不一致），判讀如上

### 回饋與修正（contract 之後）
- AC-13：移除能力的說明斷言補上「這份繪圖板上沒有的編號也整次拒絕」。
- NFR-2：補一個斷言：三件繪圖能力都不另設等待上限。
