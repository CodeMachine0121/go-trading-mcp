# Architecture — 助理認得奇異筆、但不畫奇異筆

## Design Goal

只改助理讀的說明文字：`trading_get_chart_drawings` 介紹 `kind` 為 `brush` 的繪圖，`trading_add_chart_drawings` 寫明不能新增 `brush`。
規則判斷全在交易服務；外掛照舊原樣轉送 `drawings`、原話帶回拒絕。

## Change Scope

| 元件 | 動作 | 原因 |
| :--- | :--- | :--- |
| `cmd/server/tool_catalog_chart_drawing.go` | 修改 | 讀取說明補一段 brush（錨點是一連串點、openTime 可能不是 K 線起始時間）；新增說明與 `drawings` 參數說明補「brush 只能使用者手畫、助手不能新增、送了整批被拒」 |
| `cmd/server/tool_catalog_chart_drawing_test.go` | 修改 | 把新句子釘進既有的說明測試；新增一個「含 brush 的一批仍原樣轉送」案例 |
| `README.md`、`.sdd/UL-MAP.md` | 修改 | 文件同步（UL-MAP 已於 PRD 階段補「奇異筆」） |

**刻意不動：** 能力清單（`tool_catalog_test.go` 的名單不變）、`internal/` 任何程式（請求組裝、轉送、錯誤轉達）、伺服器開場說明（已涵蓋「只能動助手畫的」）。外掛**不**在本地擋 brush——擋了等於重複交易服務的規則，規則改了兩邊會漂移。

## New Components

無。只新增一個套件內字串常數 `chartDrawingBrushNote`（讀取說明用），與既有 `chartDrawingAssistantOnlyNote` 同一手法。

## Axis of Change

下一個可能的需求是「再多一種只有使用者能畫的種類」或「讓助手也能畫 brush」。兩者都只改這一檔的說明字串與對應測試片語；轉送路徑不需動，因為外掛從不解析 `drawings`。

## Traceability

| PRD 情境 | 元件 / 驗證 |
| :--- | :--- |
| US-01 看繪圖板的說明介紹奇異筆 | 讀取說明含 `brush`、「一連串點」— `TestChartDrawingAbilitiesTeachTheAssistantToDrawOnlyItsOwnLines` |
| US-01 點時間不一定是 K 線起始時間 | 讀取說明含「不一定是 K 線的起始時間」— 同上 |
| US-02 可畫的種類照舊 | `TestAddingDrawingsSpellsOutEveryAllowedValue`（八種與點數不變） |
| US-02 說明寫明不能畫奇異筆 | 新增說明與 `drawings` 說明含 `brush`、「只能由使用者手畫」— 同上兩個測試 |
| US-02 仍送了奇異筆時照原話轉達拒絕 | `TestAddingDrawingsHandsTheIntervalAndEveryDrawingOverUntouched` 新案例（brush 原樣轉送）＋既有的拒絕原話轉達行為 |
