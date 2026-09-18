# 策略機器人的部位規劃 — Contract Verification Matrix

**Contract source:** `.sdd/2026-09-18-strategy-bot-position-plan/PRD.md`（Acceptance Criteria 為 oracle）
**Design map:** `.sdd/2026-09-18-strategy-bot-position-plan/ARCH.md`
**Glossary:** 交易服務的 `.sdd/UL-MAP.md`（那七個詞由它定義）
**Verified:** 2026-09-18
**Ceiling:** 靜態一致性稽核。逐條把**測試斷言**與**宣告本身**各自對照規格推出的 oracle，
不以「跑完全套變綠」當判準，也不自行發明並執行新的情境。

---

## Clauses

### US-01 — 助理組得出一台會建議部位的機器人

| ID | Clause | Oracle（由規格推出） | Implementation | Test | Test audit | Code audit | Status |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| AC-01 | 建立時給整組 | 那個物件原樣送到交易服務 | `tool_catalog_automation.go:28`（`bodyParameter` ⇒ `ToolParameterLocationBody`） | `tool_catalog_test.go:265`（`trading_create_strategy_bot` 那一輪） | asserts-oracle | produces-oracle | ✅ conforms |
| AC-02 | 修改時也給得出來 | 同上 | 同上——改寫吃的是同一份 `strategyBotWriteParameters()` | `tool_catalog_test.go:265`（`trading_update_strategy_bot` 那一輪） | asserts-oracle | produces-oracle | ✅ conforms |
| AC-03 | 兩支共用同一份欄位清單 | 部位規劃在兩支上一字不差；改寫多出來的只有路徑上的識別碼 | `tool_catalog_automation.go` 兩處都呼叫同一個函式（**未改動**） | `tool_catalog_test.go:265`（table-driven 跑兩支，**同一組斷言**） | asserts-oracle | produces-oracle | ✅ conforms |
| AC-04 | 完全沒提部位規劃 | 那個欄位不進 body；那台機器人不建議部位 | `ApiToolDomain.BuildRequest`（**未改動**的選填欄位行為） | `tool_catalog_test.go:119`（切片前既有的「沒填任何欄位」那條，仍綠） | asserts-oracle | produces-oracle | 🟡 partial |
| AC-05 | 部位規劃不是必填 | `IsRequired` 為假 | `tool_catalog_automation.go:28` 末參數 `false` | `tool_catalog_test.go:265` | asserts-oracle | produces-oracle | ✅ conforms |
| AC-06 | 填了交易服務不接受的數字 | 連接器照送；交易服務整台拒絕，理由原樣回來 | 這個專案**沒有任何一處讀那五個數字**（`grep positionPlan` 只命中兩行，皆為宣告與說明） | — | no-test | produces-oracle | 🟡 partial |

**AC-04 為何 partial：** 「選填沒填就不進 body」是 `BuildRequest` 切片前就有的行為，
由 `tool_catalog_test.go:119`（對每一件能力送出空白表單）守著、未改動仍綠。

**AC-06 為何 partial：** 否定性陳述（「連接器不驗證」）。
由 Out of Scope 檢查守住：整個專案 `grep positionPlan` 只有兩行，
沒有任何一處比對取值、補預設值或拆解那個物件。

### US-02 — 說明講得出助理填不對就會出錯的每一件事

| ID | Clause | Oracle | Implementation | Test | Test audit | Code audit | Status |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| AC-07 | 整組選填，而資金是開關 | 兩句都在 | `tool_catalog_automation.go:29-38` | `tool_catalog_test.go:281-282`（`整組可以不給`、`capital 是這一組的開關`） | asserts-oracle | produces-oracle | ✅ conforms |
| AC-08 | 金額用字串給精確小數 | 有那一句 | 同上 | `tool_catalog_test.go:283` | asserts-oracle | produces-oracle | ✅ conforms |
| AC-09 | 槓桿不填即不上槓桿，而現貨不要填 | 兩句都在 | 同上 | `tool_catalog_test.go:285`（`現貨帳戶不要給`）＋說明裡「不給 leverage 就是不上槓桿」 | asserts-oracle | produces-oracle | ✅ conforms |
| AC-10 | 兩個距離是百分點，各自可單獨不填 | 兩句都在 | 同上 | `tool_catalog_test.go:286`（`百分點`）＋說明裡「各自可以單獨不給」 | asserts-oracle | produces-oracle | 🟡 partial |
| AC-11 | 五樣都被點名 | 助理知道那個物件裡放什麼 | 說明裡的形狀範例＋逐項說明 | `tool_catalog_test.go:293`（六個鍵名逐一斷言） | asserts-oracle | produces-oracle | ✅ conforms |

**AC-10 為何 partial：** 「是百分點」有斷言；「各自可以單獨不給」那一句在說明字串裡，
但沒有獨立斷言——它與「整組可以不給」同一句話的後半，而前半有斷言。

### US-03 — 助理知道那個停損沒有被回測驗證過

| ID | Clause | Oracle | Implementation | Test | Test audit | Code audit | Status |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| AC-12 | 那一支能力的說明講出這件事 | 說明講出回測沒有把止損止盈算進去 | `tool_catalog_automation.go:48`——寫在**能力**上而不是欄位上：它講的不是「這一欄怎麼填」，而是「這件事與你剛剛做過的另一件事沒有對過帳」 | `tool_catalog_test.go:302`（斷言含「回測」「止損」「沒有對過帳」） | asserts-oracle | produces-oracle | ✅ conforms |

### Core Business Rules（PRD §4）

| ID | Clause | Oracle | Implementation | Test | Test audit | Code audit | Status |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| BR-01 | 是一個巢狀物件，不是五個平行欄位 | 型別為 `Object`，一欄而非五欄 | `tool_catalog_automation.go:28`（`ToolParameterKindObject`） | `tool_catalog_test.go:275`（斷言 `Kind` 為 `object`） | asserts-oracle | produces-oracle | ✅ conforms |
| BR-02 | 選填，不補預設值 | 省略時不進 body；連接器不寫第二份預設值 | 末參數 `false`；`allIn` 這個字只出現在**說明句子**裡，不是常數也不是預設值 | `tool_catalog_test.go:277` | asserts-oracle | produces-oracle | ✅ conforms |
| BR-03 | 不驗證取值 | 認不得的值照送 | 沒有任何讀取或比對那五個數字的程式碼 | — | no-test | produces-oracle | 🟡 partial |
| BR-04 | 建立與改寫共用同一份欄位清單 | 沒有哪個欄位只在一條路上能填 | 兩處呼叫同一個函式（未改動） | `tool_catalog_test.go:265`（同一組斷言跑兩支） | asserts-oracle | produces-oracle | ✅ conforms |
| BR-05 | 說明必須講出六件事 | 六句都在 | `tool_catalog_automation.go:29-38` | `tool_catalog_test.go:281-286`（六條各一） | asserts-oracle | produces-oracle | ✅ conforms |
| BR-06 | 能力的說明必須講出回測沒算止損止盈 | 有那一段 | `:48` | `tool_catalog_test.go:302` | asserts-oracle | produces-oracle | ✅ conforms |

**BR-03 與 AC-06 同一條理由**，見上。

### Edge Cases（PRD §4）

| ID | Clause | Oracle | Implementation | Test | Test audit | Code audit | Status |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| EC-01 | 只填 `capital`，其餘四樣不填 | 照送；由交易服務讀出預設值 | 連接器不拆那個物件 | — | no-test | produces-oracle | 🟡 partial |
| EC-02 | 填了槓桿卻沒填資金 | 照送；交易服務讀作沒有部位規劃 | 同上 | — | no-test | produces-oracle | 🟡 partial |
| EC-03 | 槓桿填 1 | 照送；與不填一字不差 | 同上 | — | no-test | produces-oracle | 🟡 partial |
| EC-04 | 交易服務日後多一個部位規劃欄位 | 只改那個物件的說明 | 那一欄是 `Object` 且**不列舉內部鍵的型別** | — | no-test | produces-oracle | 🟡 partial |

**EC-01～EC-04 為何全部 partial：** 四條都是同一件事的四種說法——
**這個連接器不碰那個物件**。由「沒有任何一處讀那五個數字」守住。
EC-02 特別值得記下來：它是**交易服務也不會拒絕**的那一種錯，
所以它的守衛不在這張表的任何一列，而在那一欄的說明裡（AC-07）。

### Non-Functional（PRD §6）

| ID | Clause | Oracle | Implementation | Test | Test audit | Code audit | Status |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| NFR-01 | 既有的每一件能力一個都不變 | 清單的形狀不變 | 沒有新增、刪除或改名任何一件能力 | `tool_catalog_test.go:83`（**逐字比對整份能力清單與登入需求**，切片前既有，未改動仍綠） | asserts-oracle | produces-oracle | ✅ conforms |
| NFR-02 | 機器人那幾支既有的欄位一個都不動 | 四個必填欄位仍在、仍必填 | `strategyBotWriteParameters()` 前四欄未改動 | `tool_catalog_test.go:311`（四欄逐一斷言存在且必填） | asserts-oracle | produces-oracle | ✅ conforms |
| NFR-03 | 部位規劃的取值與預設值不在這個專案裡出現第二份 | 沒有第二份清單 | `allIn`／`percentage`／`fixedAmount` 只出現在**說明句子**裡，沒有常數、enum 或比對 | — | no-test | produces-oracle | 🟡 partial |
| NFR-04 | 每一個欄位都有說明、有型別 | 兩者皆非空 | `bodyParameter` 給了 | `tool_catalog_test.go:103`（切片前既有，對每一件能力的每一欄，仍綠） | asserts-oracle | produces-oracle | ✅ conforms |

---

## Orphans

| # | Behavior | Site | Explained by | Judgement |
| :--- | :--- | :--- | :--- | :--- |
| 1 | `TestWritingAStrategyBotStillAsksForEverythingItAlwaysDid` | `tool_catalog_test.go:311` | 無 PRD 條款（NFR-02 的落點） | ⚠️ 良性且值得留。這一刀在一份**共用**的欄位清單上動手，而共用清單最容易出的錯是順手改壞別的欄位。四欄逐一釘住必填，是那個風險唯一的守衛 |
| 2 | 那一欄的說明特別長（六句） | `tool_catalog_automation.go:29-38` | AC-07～AC-11 | ⚠️ 良性且必要。這是整份清單裡**唯一一個填錯不會被拒絕的欄位**——現貨填槓桿、漏填資金，交易服務都收下。說明是這裡唯一的守衛，而它的每一句都是驗收項 |
| 3 | 重用上一刀留下的 `abilityNamed`／`boxNamed` | `tool_catalog_test.go:159`／`:174` | 上一刀的 orphan #1 | ✅ 那兩個輔助函式現在被六條測試共用，比上一刀更站得住 |

**Out of Scope 檢查**：PRD §1 列的四項**沒有任何一項有對應程式碼**。逐一確認：

- **在連接器裡驗證那五個數字**：`grep positionPlan` 只有兩行，都不是比對。
- **在連接器裡補預設值**：`allIn` 只在說明句子裡，不是常數也不是預設值。
- **替助理決定要不要填**：`IsRequired` 為假，連接器不填任何值。
- **讀取／清單／執行紀錄那幾支的形狀**：`trading_get_strategy_bot`、
  `trading_list_strategy_bots`、`trading_list_strategy_bot_runs`
  三處說明與欄位**一個字都沒改**——它們不宣告回應欄位，
  交易服務多回 `positionPlan` 與那三個數字就自然帶到助理面前。

**`internal/` 完全未改動**（`git diff` 只碰 `cmd/server/` 兩個檔與 `.sdd/`），
這是 `ARCH.md` §3 說的那次檢驗：上一刀證明「搬一個欄位」不必動到任何一層，
這一刀證明「加一個欄位」也不必。

無越界。

---

## Summary

| Status | Count |
| :--- | :--- |
| ✅ conforms | 14 |
| 🔴 violation | 0 |
| 🟠 mis-asserted | 0 |
| 🟡 partial | 9 |
| ❌ gap | 0 |
| ❔ unclear | 0 |
| ⚠️ orphan | 2（皆良性，其中一項必要）＋1 重用 |

**Clauses:** 23 · **Conformance:** 61%（14/23 完全一致）

九條 partial 裡有七條是同一個否定性陳述的七種說法——**這個連接器不碰那個物件**
（AC-06、BR-03、EC-01～EC-04、NFR-03）；另兩條是一段切片前就被守著的既有行為（AC-04）
與一句話的後半（AC-10）。

**partial 佔比在這個專案比交易服務高，是轉達層的常態而不是品質訊號**：
一個轉達層的承諾大半是「我不做什麼」，而那些承諾的證據是**程式碼裡沒有那幾行**。
上一刀是 70%、這一刀 61%，兩次的差別只是這一刀的 Edge Cases 全部落在那個否定性陳述上。

### 值得記下來的一件事

**這是整份工具清單裡唯一一個「填錯不會被拒絕」的欄位。**

別的欄位錯了，交易服務會回一句話，而助理讀得到那句話、改得動。
這一欄錯了——替現貨帳戶填了槓桿、或填了其餘四樣卻漏了資金——
**交易服務會收下**，因為那兩種都是合法的值組合。代價落在讀訊息的那個人身上：
一則現貨訊息多出兩行不該有的槓桿，或一台他以為會建議部位、實際上什麼都不建議的機器人。

所以那一欄的說明不是文案，它是**這裡唯一的守衛**，而 `ARCH.md` §7 特別留了一句話給下一位：
改它的說明之前，先想清楚你拿掉的是哪一道。
