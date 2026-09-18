# 回測的止損止盈（連接器）— Contract Verification Matrix

**Contract source:** `.sdd/2026-09-18-backtest-exit-levels/PRD.md`（Acceptance Criteria 為 oracle）
**Design map:** `.sdd/2026-09-18-backtest-exit-levels/ARCH.md`
**Glossary:** `.sdd/UL-MAP.md`
**Scope:** `go-trading-mcp`
**Verified:** 2026-09-18
**Ceiling:** 靜態一致性稽核。**這份清單就是介面**，讀它的是一個模型——
所以「說明裡有沒有寫那句話」與「欄位存不存在」是同一個等級的契約，兩者都逐條斷言。

---

## Clauses

### US-01 — 助手說得出那兩個出場距離

| ID | Clause | Oracle | Implementation | Test | Status |
| :--- | :--- | :--- | :--- | :--- | :--- |
| AC-01 | 兩個參數皆選填 | `required=false` | `tool_catalog.go:114`、`:123` | `tool_catalog_test.go:303`（兩支各斷言 `IsRequired == false`） | ✅ conforms |
| AC-02 | 兩支共用**同一份** | 加在 `backtestParameters()` 裡 | `tool_catalog.go:113`（在共用函式內） | `tool_catalog_test.go:303`（兩支都有）＋**既有** `:232` 的逐欄 `assert.Equal` 迴圈——它比對的是**整個 `ToolParameterVo`**，所以「抄成兩份且漂了一句話」會紅 | ✅ conforms |
| AC-03 | 種類是 `string` | 與既有精確小數一致 | `vo.ToolParameterKindString` | `tool_catalog_test.go:303` | ✅ conforms |
| AC-04 | 連接器不驗證 | 沒有任何檢查 | 兩格只有 `bodyParameter(...)`，無分支 | 無專屬測試——**沒有東西可測**：沒寫出來的程式碼沒有行為 | ✅ conforms |

**AC-02 的證據最強的一半是既有那條測試。** `TestReplayingAStrategyScriptStillAsksWhichWayToTrade`
（切片前就存在）對交易策略重演的每一欄，斷言腳本重演有**一模一樣的整個值**。
這一刀什麼都不必加，那條測試就已經替「共用而非兩份」把關了——
而那正是它當初被寫成比對整個值（而非只比名字）的理由。

### US-02 — 說明講得出助手看不出來的那三件事

| ID | Clause | Oracle | Implementation | Test | Status |
| :--- | :--- | :--- | :--- | :--- | :--- |
| AC-05 | 留白＝完全不模擬 | 明寫「不是套用一個常見的預設值」 | `tool_catalog.go:117` | `tool_catalog_test.go:333`（`不給就是完全不模擬止損`） | ✅ conforms |
| AC-06 | 距離從**進場價**量 | 明寫 | `tool_catalog.go:115` | `:333`（`進場價`） | ✅ conforms |
| AC-07 | 同一根碰到兩個算止損 | 明寫，並說出為什麼 | `tool_catalog.go:126` | `:333`（`一律算止損`） | ✅ conforms |
| AC-08 | 負／超 100 拒絕，正好 100 可以 | 明寫 | `tool_catalog.go:119` | `:333`（`正好 100 可以`） | ✅ conforms |
| AC-09 | 與機器人那組同名不同事 | 明寫兩者各自是什麼 | `tool_catalog.go:120` | `:333`（`同名、不同事`） | ✅ conforms |

### US-03 — 建立機器人那句警告改成指路

| ID | Clause | Oracle | Implementation | Test | Status |
| :--- | :--- | :--- | :--- | :--- | :--- |
| AC-10 | 舊那句話改掉 | 不再出現「從頭到尾不把止損止盈算進去」 | `tool_catalog_automation.go:51` | `tool_catalog_test.go:354`（**`NotContains` 舊句子**） | ✅ conforms |
| AC-11 | 不謊稱已經算進去 | 說「只在你填進去時才算」 | `tool_catalog_automation.go:51` | 由 AC-13 的 `還沒有對過帳` 間接釘住 | 🟡 partial |
| AC-12 | 指路 | 點名那兩個欄位 | `tool_catalog_automation.go:53` | `:354`（兩個欄位名各一條 `Contains`） | ✅ conforms |
| AC-13 | 本來要防的事仍防住 | 「還沒有對過帳」 | `tool_catalog_automation.go:52` | `:354` | ✅ conforms |
| AC-14 | 位置不動 | 仍在「建立不等於啟動」之後 | `tool_catalog_automation.go:48-49` | 無專屬測試——順序不是行為 | 🟡 partial |

**AC-11 為何 partial**：沒有一條斷言直接讀「只在你填進去時才算」那個子句。
釘住的是它的**兩個後果**：舊句子不在了（AC-10）、且沒有變成無條件的保證（`還沒有對過帳` 仍在）。
逐字斷言那個子句會把措辭釘死，而那句話的**意思**已經被兩側夾住了。

**AC-14 為何 partial**：說明文字裡的段落順序沒有被斷言。測順序要比對整段字串，
而那會讓任何一次措辭改善變成一次假失敗。

### Core Business Rules（PRD §4）

| BR | Implementation | Test | Status |
| :--- | :--- | :--- | :--- |
| BR-01 兩格放共用組 | `tool_catalog.go:113` | `:303` ＋既有 `:232` | ✅ |
| BR-02 交易模式維持各自加 | **未改動** | 既有 `:215`（交易策略重演**沒有** `tradingMode`）仍綠 | ✅ |
| BR-03 不驗證 | 無程式碼 | — | ✅ |
| BR-04 不補預設值 | 無程式碼 | — | ✅ |

### Non-Functional（PRD §6）

| 要求 | 驗證 | Status |
| :--- | :--- | :--- |
| 既有能力一個都不動 | 既有 `TestTheCatalogueCoversEveryThingTheTradingServiceOffersAndNothingElse`（字面清單，未改一行）＋`TestWritingAStrategyBotStillAsksForEverythingItAlwaysDid` 全綠 | ✅ conforms |
| 精確小數一律 `string` | `:303` 斷言 `Kind` | ✅ conforms |
| 新增只動一處 | `git diff --stat`：`tool_catalog.go` 一個函式 | ✅ conforms |

---

## Orphans

| 項目 | 判斷 |
| :--- | :--- |
| `backtestParameters()` 註解多一段「屬於這裡的判準是一句問題」 | **合理**。交易模式那一段註解答的是「為什麼不在這裡」；這一段答的是「什麼才在這裡」。兩段合起來才是一條可以套用到下一個參數的判準 |
| README 多一句示範（「這次停損放 2%，差很多嗎？」） | **合理**。那句話就是這一刀存在的理由，用使用者會說的話寫出來 |
| UL-MAP 多一條歧義（同名的兩組距離） | **合理**，而且是這個地圖該收的那一類——它只收連接器自己的詞彙與**會被搞混的東西** |

規格說了而沒實作的：**無**。

---

## Summary

| | 數量 |
| :--- | ---: |
| ✅ conforms | 19 |
| 🟡 partial | 2 |
| ❌ violates | 0 |

**兩個 partial 都是「不逐字釘死措辭」的刻意選擇**，各自的意思已被兩側的斷言夾住。

### 值得記下來的一件事

**「共用還是各自加」現在有一條寫下來的判準，而它是同一個註解回答的第二個問題。**

交易模式那一刀留下的註解說的是「為什麼**不**在共用組」——因為那個端點會默默忽略它。
這一刀補上的是「什麼**才**在共用組」——同一句問題的另一個答案：
**那個端點真的會用它嗎？**

下一個重演參數（手續費率、滑點）來的時候，不必再從頭想一次。
而 `TestReplayingAStrategyScriptStillAsksWhichWayToTrade` 比對整個值而不只比名字，
是這條判準唯一的執行者。
