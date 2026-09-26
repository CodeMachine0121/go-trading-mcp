# Product Requirements Document (PRD)

**Status:** Draft
**Version:** v1.0
**Owner:** James Hsueh
**Stakeholders:** Engineering

---

## 1. Background & Goal (Why & Goal)

- **Problem Statement:** 使用者今天要把電子郵件與密碼告訴助理，密碼因此出現在對話裡；身分又只存在某一台外掛的記憶裡，
  那一台重啟或換一台，所有人都得重新登入，外掛也因此只能跑一份。
- **Expected Outcome:** 使用者在 Claude Code 的外掛選單按「連線」、在瀏覽器登入交易服務的網站並按下允許，就能使用外掛；
  外掛從此不經手任何密碼、不保管任何身分，只確認每一次帶來的**外掛授權**是真的、是給這個外掛的，然後以那份身分代辦。
  外掛可以同時跑不只一份。
- **Out of Scope:** 授權伺服器本身（發授權、登入與允許頁面）；部署設定的實際修改；權限範圍細分。

## 2. User Personas

- **Primary Role:** 透過 Claude Code 使用交易助理的系統擁有者；以及 Claude Code 本身——它負責走完瀏覽器登入、保管與續用外掛授權。
- **Usage Context:** 第一次連上外掛、外掛授權過期之後、或外掛重新部署之後。

## 3. User Stories & Acceptance Criteria

### US-01 — 沒有有效外掛授權就不接待 [priority: P0]
**As a** 系統擁有者，**I want** 外掛只接待帶著有效、發給它的外掛授權的呼叫，**so that** 沒有人能不經我允許就以任何身分使用它。

```gherkin
Scenario: 帶著有效外掛授權照常接待
  Given 帶著交易服務確認有效、發給這個外掛的外掛授權
  When 助理列出會做的事
  Then 照常列出會做的事

Scenario: 沒有帶授權就不接待
  Given 沒有帶任何外掛授權
  When 助理連上外掛
  Then 外掛不接待
  And 回覆指出這個外掛的授權說明在哪裡

Scenario: 失效的授權不接待
  Given 帶來的外掛授權被交易服務判定為失效
  When 助理查一段 K 線
  Then 外掛不接待
  And 回覆指出這個外掛的授權說明在哪裡

Scenario: 發給別的服務的授權不算數
  Given 帶來的外掛授權有效，但發給的對象是別的服務
  When 助理查一段 K 線
  Then 外掛不接待

Scenario: 對象只差結尾斜線算同一個
  Given 帶來的外掛授權有效，發給的對象是這個外掛、只是多了結尾的斜線
  When 助理查一段 K 線
  Then 照常接待

Scenario: 問不到交易服務不說成授權失效
  Given 交易服務暫時不在
  When 助理帶著外掛授權連上外掛
  Then 外掛不接待
  And 回覆說「暫時連不到交易服務」
  And 回覆不指引重新授權
```

### US-02 — 同一份授權的確認結果記一小段時間 [priority: P1]
**As a** 系統擁有者，**I want** 同一份授權不必每一件事都重新問交易服務，**so that** 一連串代辦不會多出一連串確認。

```gherkin
Scenario: 一分鐘內再出現不再確認
  Given 同一份外掛授權三十秒前確認為有效
  When 助理再做一件事
  Then 外掛不再向交易服務確認

Scenario: 滿一分鐘重新確認
  Given 同一份外掛授權六十秒前確認為有效
  When 助理再做一件事
  Then 外掛重新向交易服務確認

Scenario: 不記得比授權到期更久
  Given 同一份外掛授權剛確認為有效，而它二十秒後到期
  When 二十五秒後助理再做一件事
  Then 外掛重新向交易服務確認

Scenario: 問不到的結果不記
  Given 上一次確認時交易服務不在
  When 助理再做一件事
  Then 外掛重新向交易服務確認
```

### US-03 — 公開說明授權找誰拿 [priority: P0]
**As a** Claude Code，**I want** 從外掛自己查到授權要向誰拿，**so that** 使用者不必事先設定任何東西就能走瀏覽器登入。

```gherkin
Scenario: 查授權說明
  When Claude Code 查這個外掛的授權說明
  Then 說明寫著這個外掛的正式位址
  And 說明寫著授權要向交易服務的公開位址拿
  And 說明寫著授權放在請求的標頭裡帶來

Scenario: 帶外掛路徑的位址回同一份說明
  When Claude Code 用帶著外掛路徑的那一個位址查授權說明
  Then 回同一份說明
```

### US-04 — 代辦時以使用者的身分進行 [priority: P0]
**As a** 使用者，**I want** 每一件事都以我在瀏覽器允許的那份身分進行，**so that** 我不必再把密碼交給助理。

```gherkin
Scenario: 改行情的事帶著授權去問
  Given 使用者已授權
  When 助理把一個標的加進觀察清單
  Then 這一次詢問帶著他這一次帶來的外掛授權
  And 交易服務的回覆原樣帶回

Scenario: 看行情的事也帶著授權去問
  Given 使用者已授權
  When 助理查一段 K 線
  Then 這一次詢問帶著他這一次帶來的外掛授權

Scenario: 交易服務不認得身分時請他重新連線
  Given 交易服務說不認得這份身分
  When 助理做任何一件事
  Then 回覆請使用者到 Claude Code 的外掛選單重新連線
  And 外掛不重試

Scenario: 帳號還沒開通時原話帶回
  Given 使用者已授權、帳號還沒開通
  When 助理建立一支策略腳本
  Then 交易服務說尚未開通的原話原樣帶回
```

### US-05 — 拿掉的能力與說明不再出現 [priority: P0]
**As a** 使用者，**I want** 外掛不再提供對話裡的登入方式，**so that** 助理不會再向我要密碼。

```gherkin
Scenario: 能力清單裡沒有對話裡的登入
  When 助理列出會做的事
  Then 清單裡沒有登入、登出、立即換新登入、建立使用者
  And 清單裡有「我是誰」與「改密碼」

Scenario: 使用說明不再要密碼
  When 助理讀外掛的使用說明
  Then 說明不再請它向使用者要電子郵件與密碼
  And 每一件能力的說明後面不再附「請先登入」
```

## 4. Business Flow & Logic

1. Claude Code 連上外掛，沒有外掛授權 → 外掛不接待，指出授權說明在哪裡。
2. Claude Code 讀授權說明 → 找到交易服務（授權伺服器）→ 帶使用者走瀏覽器登入與允許 → 拿到外掛授權。
3. 之後每一次呼叫都帶著外掛授權：外掛向交易服務確認（有效、使用者還在、發給這個外掛），結果記一小段時間。
4. 代辦時把同一份外掛授權轉給交易服務；交易服務不認得 → 請使用者到外掛選單重新連線；其他拒絕原話帶回。
5. 外掛授權過期時由 Claude Code 自己續用，外掛不參與。

## 5. UI/UX Design & Interaction

- 沒有畫面。使用者看到的是 Claude Code 外掛選單的「連線」與交易服務網站的登入與允許頁面（網站那一刀）。

## 6. Non-Functional Requirements

- 安全：外掛不經手密碼、不保管任何身分；外掛授權絕不出現在回覆或紀錄裡；確認結果只以授權的指紋記住，不記授權本身。
- 可用性：外掛可同時跑不只一份，任何一份重啟都不必重新授權。
- 效能：同一份授權一分鐘內至多確認一次。

## 7. Dependencies & Risks

- **部署順序**：必須在交易服務的授權伺服器上線、且部署設定（外掛的公開位址與交易服務的公開位址）完成之後才部署；
  反過來則所有人都無法使用外掛。
- 交易服務確認授權的能力若變慢，每一分鐘的第一件事會跟著變慢。

## 8. Appendix

- 跨專案合約：外掛是受保護的資源、交易服務是授權伺服器、網站提供登入與允許頁面。
- 代替提問所做的決定見 BRIEF 的 Open Decisions。
