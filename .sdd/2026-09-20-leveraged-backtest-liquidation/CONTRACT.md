# 回測的槓桿與強制平倉（連接器）— Contract Conformance Matrix

**Oracle:** `PRD.md` §3 Acceptance Criteria（13 個情境）、§4 Core Business Rules（4 條）、§6 Non-Functional Requirements（3 條）
**Audited:** 2026-09-20
**Ceiling:** 這是一次**靜態一致性稽核**。每一條都先只從規格推出預期結果，再**各自獨立**判斷
「測試有沒有斷言這個結果」與「程式有沒有產出這個結果」。它不撰寫新探針、不執行自己發明的情境。

**這一刀的特殊之處：產出幾乎全是文字。** 助手讀得到的只有欄位說明，
所以「程式有沒有產出 oracle」在多數條款上等於「那段話有沒有說出那件事」。
這使得**斷言得夠精確**特別要緊——一個只驗關鍵字的斷言，放得過一段被改壞的說明。
本次稽核抓到的兩件事都屬於這一類。

---

## Clauses

### US-01 — 兩支重演都問得出槓桿

| ID | 情境 | Oracle | 實作 | 測試 | T | C | 狀態 |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| AC-01 | 重演一支策略腳本說得出槓桿 | 有槓桿倍數與維持保證金率兩格，皆選填 | `tool_catalog.go` `backtestParameters()` | `TestBothReplaysTakeTheSameLeverage`（含 `IsRequired=false`、型別為字串） | asserts-oracle | produces-oracle | ✅ |
| AC-02 | 重演一份交易策略說得出同樣兩格 | 同樣兩格，**說明一字不差** | 同上（兩支共用同一組） | `TestBothReplaysTakeTheSameLeverage` ＋ `TestBothReplaysWordTheLeverageIdentically`（**本次補上維持保證金率那一格**） | asserts-oracle | produces-oracle | ✅ ¹ |
| AC-03 | 槓桿與交易模式恰恰相反 | 那一支沒有交易模式那一格，但有槓桿兩格 | `tool_catalog_strategy.go`（`tradingMode` 只在策略腳本那一支） | 既有的 `TestReplayingATradingStrategyHasNoTradingModeToGive` ＋ `TestBothReplaysTakeTheSameLeverage` | asserts-oracle | produces-oracle | ✅ |
| AC-04 | 填了的槓桿真的到得了交易服務 | 送出去的內容裡帶著槓桿倍數 5 | 宣告本身 ＋ `ToolArgumentsDomain.EncodedSubset`（既有） | `TestAFilledInLeverageActuallyLeavesTheConnector` | asserts-oracle | produces-oracle | ✅ |
| AC-05 | 兩格都不填時送出去的內容一字不差 | 內容裡沒有那兩個名字 | `EncodedSubset`（既有：只收填過的名字） | `TestLeavingTheLeverageOutSendsWhatItAlwaysSent` | asserts-oracle | produces-oracle | ✅ |

¹ 稽核前只驗了 `leverage` 一格，比 oracle 弱——`maintenanceMarginRate` 的說明在兩支之間漂移不會被抓到。已改成兩格都驗。

### US-02 — 說明寫出送出去之前查不到的事

| ID | 情境 | Oracle | 測試 | T | C | 狀態 |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| AC-06 | 說得出留白是什麼意思 | 說出不給／0／1 都是不借錢，也不模擬強制平倉 | `TestTheLeverageSaysWhatCannotBeDiscoveredBySending`（逐字） | asserts-oracle | produces-oracle | ✅ |
| AC-07 | 說得出撐得住多遠，並給實際數字 | 說出約 `100÷槓桿` 個百分點，且至少一個實際數字 | 同上——**三個各自唯一的片語**：約略式、精確式、`5 倍約 19.5%` | asserts-oracle | produces-oracle | ✅ ² |
| AC-08 | 說得出止損比強平近時止損先出場 | 說出開了槓桿就該一起給止損距離 | 同上（`stopLossPercentage`；整句被刪的突變會紅） | asserts-oracle | produces-oracle | ✅ |
| AC-09 | 說得出會被拒絕的那幾種 | 說出現貨給大於 1 會整次被拒絕 | 同上（逐字），**另加「介於 0 與 1 之間會整次被拒絕」** | asserts-oracle | produces-oracle | ✅ ⁶ |
| AC-10 | 維持保證金率留白不是關掉它 | 說出留白時是 0.5%，且與隔壁不一樣 | `TestTheMaintenanceMarginSaysItsBlankIsNotTheOthers`（**兩句各驗一半**：「不給不是關掉它，是用 0.5%」與「這與旁邊每一格的留白都不一樣」） | asserts-oracle | produces-oracle | ✅ ⁵ |
| AC-11 | 說得出上限是算出來的 | 說出必須小於 `100÷槓桿` | 同上 | asserts-oracle | produces-oracle | ✅ |

² 稽核前只驗 `100÷槓桿`——而那個字串在那段說明裡出現**兩次**（約略式與精確式），
刪掉任一個測試照樣綠。**這一條是突變測試抓到的唯一倖存者**，已改成三個互不重疊的片語。

### US-03 — 成績單多的那一格要看得懂

| ID | 情境 | Oracle | 實作 | 測試 | T | C | 狀態 |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| AC-12 | 重演一支策略腳本說得出那一格 | 說出有強平出場筆數，並說出為什麼非看不可 | `liquidationReportCardNote` | `TestBothReplaysSayWhyTheWipeOutCountMatters`（`liquidationExitCount` ＋「歸零過三次」） | asserts-oracle | produces-oracle | ✅ |
| AC-13 | 重演一份交易策略說得出同一件事，措辭一字不差 | 同一段話，逐字相同 | 同一個常數 | 同上，**本次補上「兩支都含有那個常數本身」** | asserts-oracle | produces-oracle | ✅ ³ |

³ 稽核前只驗「兩支都提到那兩個詞」，比 oracle 弱：一支把段落抄一份、改掉幾句但留著那兩個詞，
測試照樣綠，而兩支就會對同一張成績單講兩種話。已補上對常數本身的斷言，
並以一個「抄成近似版」的突變確認它會紅。

### Core Business Rules

| ID | 規則 | Oracle | 實作 | T | C | 狀態 |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| BR-1 | 兩格放在兩支共用的那一組參數裡 | 兩支都有，不是各自加一次 | `backtestParameters()` | asserts-oracle | produces-oracle | ✅ |
| BR-2 | 連接器不驗證 | 這兩格不會在連接器被拒絕 | 沒有新增任何判斷 | asserts-oracle | produces-oracle | ✅ ⁴ |
| BR-3 | 留白的不上線 | 沒填的名字不出現在內文裡 | `EncodedSubset`（既有） | asserts-oracle | produces-oracle | ✅ |
| BR-4 | 成績單那段說明兩支共用一份 | 兩支的那一段逐字相同 | `liquidationReportCardNote` | asserts-oracle | produces-oracle | ✅（見註 ³） |

⁵ 稽核當下誤判為「一句涵蓋兩半」，由外部 code review 指出。已補第二個斷言。
⁶ 原本只說了現貨那一種拒絕；`0 < x < 1` 那一種是 code review 補上的。

⁴ 連接器對一個欄位**唯一做得到**的判斷就是必不必填，而那一項被斷言為「選填」。
再多的驗證都得是新程式碼，不是既有路徑的分支——所以這一條的守衛是那個 `IsRequired=false`，
加上 Out of Scope 的明文。

### Non-Functional Requirements

| ID | 要求 | 判定 | 狀態 |
| :--- | :--- | :--- | :--- |
| NFR-1 | 兩格都不填時送出去的內容一字不差 | 同 AC-05／BR-3 | ✅ |
| NFR-2 | 兩支重演對這兩格的說明逐字相同 | 同 AC-02（兩格都驗）與 AC-13 | ✅ |
| NFR-3 | 不因這一刀新增任何驗證 | 同 BR-2 | ✅ |

---

## Orphans

| 行為 / 符號 | 說明 | 判定 |
| :--- | :--- | :--- |
| `apiToolNamed` 測試輔助 | 取能力本身而非它的對外形狀，只為了問一個形狀答不了的問題：填進去的那一格**真的離得開連接器嗎**。 | 非孤兒（AC-04／AC-05） |
| `liquidationReportCardNote` 的位置 | `/improve-codebase` 階段從 `tool_catalog.go` 搬到它兄弟 `costedReportCardNote` 旁邊。行為零改動。 | 非孤兒（結構性） |

**Out of Scope 反向檢查**（皆未實作，無越界）：連接器自行驗證任何一格、資金費率、
機器人 `positionPlan` 裡那一格槓桿。

---

## Summary

| | 數量 |
| :--- | ---: |
| ✅ conforms | 20 |
| 🔴 violations | 0 |
| 🟠 mis-asserted | 0（稽核前 2，已修） |
| 🟡 partial | 0 |
| ❌ gaps | 0 |
| ❔ unclear | 0 |
| ⚠️ orphans | 0 |

**Conformance: 20 / 20 ＝ 100%**

### Code review 之後補上的四件事（`ad256a6` 之後）

外部 review 抓到四個 LOW，四個都是這一刀造成的，四個都已修：

1. **`exitReason` 被我加成出現兩次。** 新的成績單說明尾端也提到它，於是既有的
   `TestReplayingAScriptSaysWhyTheStopCountMatters` 停止咬人——止損那一段可以整句
   掉光而測試照樣綠。改成釘住那一整句（`每一筆交易自己也帶著 exitReason`）。
2. **AC-10 只驗了一半。** 測試名字與這份文件都說它涵蓋兩半，實際上只驗了
   「不給不是關掉它，是用 0.5%」；那句對照子句刪掉照樣綠。已補上第二個斷言，
   本表的 AC-10 一列也一併更正——**那句「同時涵蓋兩半」原本就是錯的**。
3. **`0 < x < 1` 會被拒絕，而說明從沒提過。** 上一句才說「0 是不借錢、1 是不借錢」，
   讀起來就像「小於等於 1 都無害」，而 0.5 正是助手想押半個部位時會打的數字。
   這個 repo 自己的規矩是「說明要說出什麼會被拒絕」，而它只說了現貨那一種。已補。
4. **手續費的基準前後矛盾。** `entryCostPercentage` 說收「押注金額」的百分比，
   新的槓桿那一格說「手續費照放大後的曝險金額算」——開了槓桿這是兩個數字。
   後端在同一刀改成收曝險，所以**舊的那句已經是錯的**。已改，並說出沒開槓桿時兩者相同。

四個各有一條斷言，並以突變確認會紅。

### 稽核當下抓到並已修正的兩件事

兩件都是**斷言比 oracle 弱**，而且都屬於同一類——這一刀的產出幾乎全是文字，
而驗文字最容易寫出「看起來有驗、其實放得過」的斷言：

1. **「兩支說明一字不差」只驗了兩格中的一格。** `maintenanceMarginRate` 的說明
   在兩支之間漂移不會被抓到。已改成兩格都驗。
2. **「兩支都說得出成績單那一格」只驗了兩個關鍵字。** 一支把段落抄一份、
   改掉幾句但留著那兩個詞，測試照樣綠——而那正是這段話存在要防的事。
   已補上對共用常數本身的斷言，並以「抄成近似版」的突變確認它會紅。

另有一件由突變測試（而非本稽核）抓到：**撐得住多遠**那一條原本只驗 `100÷槓桿`，
而那個字串在說明裡出現兩次，刪掉任一個都不會紅。已改成三個互不重疊的片語。
