# 加入市集的策略腳本是拿到一份副本（外掛） — Architecture Design

**Status:** Confirmed · **Source PRD:** `PRD.md`

| Area | Action | What / Why |
| :--- | :--- | :--- |
| `cmd/server/tool_catalog_strategy.go` | **Modify** | 加入、列出、讀取、改寫腳本、上架／下架、建立／改寫交易策略的說明；移除 `trading_abandon_strategy_script` |
| `cmd/server/tool_catalog_test.go` | **Modify** | 工具清單不再含放棄採用；說明提到複製與只能用自己的 |
| 其他 | **Not touched** | 工具都是對交易服務的直通，沒有自己的規則 |
