# 交易策略的交易模式 — Contract Verification Matrix

**Contract source:** `.sdd/2026-09-18-trading-strategy-trading-mode/PRD.md`（Acceptance Criteria 為 oracle）
**Design map:** `.sdd/2026-09-18-trading-strategy-trading-mode/ARCH.md`
**Glossary:** 交易服務的 `.sdd/UL-MAP.md`（交易模式三個詞由它定義）
**Verified:** 2026-09-18
**Ceiling:** 靜態一致性稽核。逐條把**測試斷言**與**宣告本身**各自對照規格推出的 oracle，
不以「跑完全套變綠」當判準，也不自行發明並執行新的情境。

---

## Clauses

### US-01 — 建立與修改交易策略時說得出交易模式

| ID | Clause | Oracle（由規格推出） | Implementation | Test | Test audit | Code audit | Status |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| AC-01 | 建立一份現貨型交易策略 | 那個值原樣出現在送往交易服務的 body 裡 | `tool_catalog_strategy.go:95`（`bodyParameter` ⇒ `ToolParameterLocationBody`） | `tool_catalog_test.go:199`（`trading_create_trading_strategy` 那一輪） | asserts-oracle | produces-oracle | ✅ conforms |
| AC-02 | 改掉一份交易策略的交易模式 | 同上 | 同上——改寫吃的是同一份 `tradingStrategyWriteParameters()` | `tool_catalog_test.go:199`（`trading_update_trading_strategy` 那一輪） | asserts-oracle | produces-oracle | ✅ conforms |
| AC-03 | 兩支共用同一份欄位清單 | 交易模式在兩支上一字不差；改寫多出來的只有路徑上的識別碼 | `tool_catalog_strategy.go:121`／`:139`（兩處都呼叫同一個函式，未改動） | `tool_catalog_test.go:199`（table-driven 跑兩支，**同一組斷言**） | asserts-oracle | produces-oracle | ✅ conforms |
| AC-04 | 完全沒提交易模式 | 那個欄位不出現在 body 裡，由交易服務套用它的預設值 | `ApiToolDomain.BuildRequest`（未改動）——選填欄位沒填就不進 body | `tool_catalog_test.go:119`（切片前既有的「沒填任何欄位」那條，仍綠） | asserts-oracle | produces-oracle | 🟡 partial |
| AC-05 | 交易模式不是必填 | `IsRequired` 為假 | `tool_catalog_strategy.go:95` 末參數 `false` | `tool_catalog_test.go:216` | asserts-oracle | produces-oracle | ✅ conforms |
| AC-06 | 填了認不得的值 | 連接器照送不自己擋；交易服務整份拒絕，理由原樣回來 | 這個專案**沒有任何一處列舉交易模式的取值**（`grep tradingMode` 只命中三行，全是宣告與說明） | — | no-test | produces-oracle | 🟡 partial |
| AC-07 | 說明講得出我該怎麼挑 | 兩個拼法的意思、不給會怎樣、不能放空時給哪一個 | `tool_catalog_strategy.go:95-101` | `tool_catalog_test.go:208-213`（四條：`spot`、`longShort`、「省略即 longShort」、「不能放空」） | asserts-oracle | produces-oracle | ✅ conforms |

**AC-04 為何 partial：** 「選填沒填就不進 body」是 `BuildRequest` 切片前就有的行為，
由 `tool_catalog_test.go:119` 那條（對每一件能力送出空白表單）守著、未改動仍綠。
沒有專門針對 `tradingMode` 的一條——它會與既有那條走完全相同的程式路徑。

**AC-06 為何 partial：** 這是否定性陳述（「連接器不驗證」）。
由 Out of Scope 檢查守住：整個專案 `grep tradingMode` 只有三行，
沒有任何一處比對取值、補預設值或轉換大小寫。
硬寫一個「它沒有擋」的測試，只會得到一個永遠綠的斷言。

### US-02 — 重演一份交易策略時沒有交易模式可以給

| ID | Clause | Oracle | Implementation | Test | Test audit | Code audit | Status |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| AC-08 | 那件能力沒有交易模式這一欄 | 欄位清單裡沒有 `tradingMode` | `tool_catalog.go:93` 的共用清單已移出它；`tool_catalog_strategy.go:193` 只吃那份共用清單，不附加 | `tool_catalog_test.go:228`（斷言**沒有**那一欄） | asserts-oracle | produces-oracle | ✅ conforms |
| AC-09 | 說明講出為什麼沒有 | 交易模式取自那一份交易策略，這裡不必也不能再說一次 | `tool_catalog_strategy.go:184` | `tool_catalog_test.go:234`（斷言說明含「交易模式」） | asserts-oracle | produces-oracle | ✅ conforms |
| AC-10 | 說明講出要改的話去哪裡改 | 指出去 `trading_update_trading_strategy` 改那一份的 `tradingMode` | `tool_catalog_strategy.go:186-187` | `tool_catalog_test.go:235-236`（斷言說明含 `trading_update_trading_strategy` 與「不能放空」） | asserts-oracle | produces-oracle | ✅ conforms |

### US-03 — 重演一支策略腳本那條路一個字都沒變

| ID | Clause | Oracle | Implementation | Test | Test audit | Code audit | Status |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| AC-11 | 那件能力仍然有交易模式，選填，說明一字不差 | 有那一欄、`IsRequired` 為假、說明與切片前相同 | `tool_catalog_strategy.go:177`——**說明字串逐字搬過來**，一個字都沒改 | `tool_catalog_test.go:245-248` | asserts-oracle | produces-oracle | ✅ conforms |
| AC-12 | 該共用的每一欄仍然共用 | 除了 `tradingMode` 之外，回測共通的每一欄在兩支上一字不差 | 兩支仍然都吃 `backtestParameters()`（`:165`、`:193`） | `tool_catalog_test.go:250-262`（逐欄比對：交易策略那一支的每一欄，除路徑識別碼外，腳本那一支都有） | asserts-oracle | produces-oracle | ✅ conforms |

### Core Business Rules（PRD §4）

| ID | Clause | Oracle | Implementation | Test | Test audit | Code audit | Status |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| BR-01 | `tradingMode` 只出現在三處 | 建交易策略、改交易策略、重演一支策略腳本 | `grep tradingMode` 命中兩個 `bodyParameter`（`:95` 供建立與改寫共用、`:177` 給腳本回測）＋一句說明文字（`:187`） | `tool_catalog_test.go:199`（兩支有）、`:228`（一支沒有）、`:245`（一支有） | asserts-oracle | produces-oracle | ✅ conforms |
| BR-02 | 重演一份交易策略沒有它 | 那一份自己記著 | `tool_catalog_strategy.go:193` 不附加 | `tool_catalog_test.go:228` | asserts-oracle | produces-oracle | ✅ conforms |
| BR-03 | 選填，不補預設值 | 省略時不進 body；連接器不寫第二份預設值 | 末參數 `false`；沒有任何一處寫 `longShort` 當預設值（那個字只出現在說明句子裡） | `tool_catalog_test.go:216`、`:247` | asserts-oracle | produces-oracle | ✅ conforms |
| BR-04 | 不驗證取值 | 認不得的值照送 | 沒有任何比對取值的程式碼 | — | no-test | produces-oracle | 🟡 partial |
| BR-05 | 建立與改寫共用同一份欄位清單 | 沒有哪個欄位只在一條路上能填 | `tool_catalog_strategy.go:121`／`:139`（未改動） | `tool_catalog_test.go:199`（同一組斷言跑兩支） | asserts-oracle | produces-oracle | ✅ conforms |
| BR-06 | 兩支回測仍然共用真的共用的那幾欄 | 拆開的只有 `tradingMode` | `backtestParameters()` 仍然是兩支的來源 | `tool_catalog_test.go:250` | asserts-oracle | produces-oracle | ✅ conforms |
| BR-07 | 說明是助理唯一的依據 | 講出兩個拼法、省略的後果、不能放空對應哪一個 | `tool_catalog_strategy.go:95-101` | `tool_catalog_test.go:208-213` | asserts-oracle | produces-oracle | ✅ conforms |

**BR-04 與 AC-06 同一條理由**，見上。

### Edge Cases（PRD §4）

| ID | Clause | Oracle | Implementation | Test | Test audit | Code audit | Status |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| EC-01 | 助理在重演一份交易策略時硬塞 `tradingMode` | 那個欄位不在清單裡，連接器不會把它放進 body | `BuildRequest` 只送**宣告過的**欄位（未改動的既有行為） | `tool_catalog_test.go:228`（那一欄不存在）＋`:119`（切片前既有，守著 `BuildRequest` 的行為） | asserts-oracle | produces-oracle | 🟡 partial |
| EC-02 | 填了空字串 | 照送；交易服務讀作省略 | 連接器不轉換 | — | no-test | produces-oracle | 🟡 partial |
| EC-03 | 填了大小寫不同的拼法 | 照送；交易服務的寬容度處理它 | 同上 | — | no-test | produces-oracle | 🟡 partial |
| EC-04 | 交易服務日後多一種交易模式 | 只改欄位說明那一句 | 三處宣告都**不列舉取值**（`Kind` 是 `string`，不是 enum） | — | no-test | produces-oracle | 🟡 partial |

**EC-01 為何 partial：** 「不在清單裡的欄位到不了交易服務」是 `BuildRequest`
切片前就有的行為。這裡新增的是「那一欄不在清單裡」，而那一半有斷言。

**EC-02／EC-03／EC-04 為何 partial：** 三條都是同一件事的三種說法——
**這個連接器不碰那個值**。由「沒有任何一處列舉取值」守住。

### Non-Functional（PRD §6）

| ID | Clause | Oracle | Implementation | Test | Test audit | Code audit | Status |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| NFR-01 | `trading_backtest_strategy_script` 的欄位與說明一字不差 | 逐字 | `tool_catalog_strategy.go:177` 的說明字串**逐字搬過來** | `tool_catalog_test.go:245`、`:250`；切片前既有的每一條仍綠 | asserts-oracle | produces-oracle | ✅ conforms |
| NFR-02 | 既有的每一件能力（名稱、需不需要登入、路徑、動詞）一個都不變 | 清單的形狀不變 | 沒有新增、刪除或改名任何一件能力 | `tool_catalog_test.go:83`（**逐字比對整份能力清單與它們的登入需求**，切片前既有，未改動仍綠） | asserts-oracle | produces-oracle | ✅ conforms |
| NFR-03 | 交易模式的取值與預設值不在這個專案裡出現第二份 | 沒有第二份清單 | `longShort`／`spot` 只出現在三處**說明句子**裡，沒有任何常數、enum 或比對 | — | no-test | produces-oracle | 🟡 partial |
| NFR-04 | 每一個欄位都有說明、有型別 | 兩者皆非空 | `bodyParameter` 三處都給了 | `tool_catalog_test.go:103`（切片前既有，對每一件能力的每一欄，仍綠） | asserts-oracle | produces-oracle | ✅ conforms |

---

## Orphans

| # | Behavior | Site | Explained by | Judgement |
| :--- | :--- | :--- | :--- | :--- |
| 1 | 新增三個測試輔助函式 `abilityNamed`／`boxNamed`／`boxNames` | `tool_catalog_test.go:160`／`:175`／`:186` | 無 PRD 條款 | ⚠️ 良性。只服務測試，且**三個都被兩條以上的測試共用**（`abilityNamed` 被三條用、`boxNamed` 被三條、`boxNames` 被一條測試的兩處）。沒有它們，每一條測試都要在讀者面前走過整份清單 |
| 2 | `backtestParameters()` 的函式名稱（「一次重演照哪些帳戶條件走」）讀起來像應該含交易模式 | `tool_catalog.go:80-93` | 由 `ARCH.md` §7 記載 | ⚠️ 良性但**必須留著那段註解**。註解寫明為什麼不含，以及統一回去的代價（其中一支會多一個被安靜忽略的假欄位） |
| 3 | 移除了共用回測清單裡的 `tradingMode` | `tool_catalog.go:93` | AC-08、BR-02 | ✅ 正確的移除。它服務的那個 endpoint 已經不收它 |

**Out of Scope 檢查**：PRD §1 列的五項全數**沒有對應程式碼**。逐一確認：

- **在連接器裡驗證交易模式**：沒有任何比對取值的程式碼。
- **在連接器裡補預設值**：`longShort` 這個字只出現在說明句子裡，不是常數也不是預設值。
- **替助理挑一個**：`IsRequired` 為假、連接器不填任何值。
- **交易策略清單／讀取工具**：`tool_catalog_strategy.go:125`／`:130` 兩處說明
  **一個字都沒改**——它們不宣告回應欄位，交易服務多回一個 `tradingMode`
  就自然帶到助理面前，這正是 PRD 判斷「這裡不必改」的理由。
- **任何機器人工具**：`strategyBotApiTools()` 與 `strategyBotWriteParameters()` 完全未改動。

**`internal/` 完全未改動**（`git diff` 只碰 `cmd/server/` 三個檔與 `.sdd/`），
這正是 `ARCH.md` §3 說的那次檢驗：搬一個欄位不該動到任何一層。

無越界。

---

## Summary

| Status | Count |
| :--- | :--- |
| ✅ conforms | 16 |
| 🔴 violation | 0 |
| 🟠 mis-asserted | 0 |
| 🟡 partial | 7 |
| ❌ gap | 0 |
| ❔ unclear | 0 |
| ⚠️ orphan | 2（皆良性）＋1 正確的移除 |

**Clauses:** 23 · **Conformance:** 70%（16/23 完全一致）

七條 partial 集中在同一件事上：**這個連接器不碰那個值**（AC-06、BR-04、EC-02、
EC-03、EC-04、NFR-03）以及**一段切片前就被守著的既有行為**（AC-04、EC-01 的另一半）。
前六條是同一個否定性陳述的六種說法，由「整個專案 `grep tradingMode` 只有三行、
沒有任何一處列舉取值」守住；後者由 `tool_catalog_test.go:119` 未改動仍綠守住。
硬寫測試只會多出六個永遠綠、永遠不會失敗的斷言。

**partial 佔比比交易服務那一刀高，是這種專案的常態而不是品質訊號**：
一個轉達層的承諾大半是「我不做什麼」——不驗證、不轉換、不補值、不列舉。
那些承諾的證據是**程式碼裡沒有那幾行**，不是測試裡有那幾條斷言。

### 值得記下來的一件事

**拿掉一個旋鈕，助理只知道自己沒有它；它不會自動知道旋鈕在哪。**

第一版只是把 `tradingMode` 從那一支移掉。那樣做的話，使用者說「我的帳戶不能放空」，
助理手上沒有可以動的東西，最可能的反應是回報它辦不到——
而正確的反應是去改那一份交易策略。

所以 `trading_backtest_trading_strategy` 的說明多的那一句
（「用 `trading_update_trading_strategy` 把那一份的 `tradingMode` 改成 `spot`」）
不是文案，是**這兩件事之間唯一的連接**，並且被 AC-10 釘住。
少了它，這一刀會讓助理從「填了沒反應」變成「知道自己辦不到」——
兩種都是壞的，只是後者比較安靜。
