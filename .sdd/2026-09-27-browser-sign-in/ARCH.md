# 在瀏覽器登入外掛 — Architecture Design

**Status:** Confirmed
**Source PRD:** `.sdd/2026-09-27-browser-sign-in/PRD.md`
**Tech context:** Go 1.26 · modelcontextprotocol/go-sdk v1.8.0（Streamable HTTP、`auth` 套件）· Onion 分層

---

## 1. Design Goal & Guiding Principle

- **In one sentence:** 讓外掛成為 OAuth 受保護資源：整個 MCP 端點前面掛一道外掛授權確認（向交易服務的 introspection 確認、以 SHA-256 指紋短暫記住），
  代辦時把呼叫端帶來的同一份 bearer 轉給交易服務；拿掉記憶體裡的身分保管與它的所有配件。
- **Guiding principle:** 「這份授權算不算數」是**一個** Domain Model 的行為（`ConnectorAuthorizationDomain`），
  「要不要再問一次」是**一個** repository 的責任（`ConnectorAuthorizationVerdictRepository`），HTTP 那一面全交給 SDK 的
  `auth.RequireBearerToken` 與 `auth.ProtectedResourceMetadataHandler`。下一個規則（例如要求某個 scope、改變記住多久）只落在其中一處。

## 2. Change Scope

| Area | Action | What / Why |
| :--- | :--- | :--- |
| `internal/domain/interface/i_trading_service_proxy.go` | **Modify** | 拿掉 `SignIn` / `RenewSession` / `RevokeSession`；加 `InspectConnectorAuthorization(ctx, accessToken)`。確認授權也是對交易服務的呼叫，照「一個外部資源一個 Proxy」放同一個介面 |
| `internal/domain/interface/i_connector_authorization_verdict_repository.go` | **Add** | 記住確認結果：`Find(accessToken, now)`、`Save(accessToken, verdict, rememberedUntil)` |
| `internal/domain/models/vo/connector_authorization_inspection_vo.go` | **Add** | 交易服務對一份授權的原始判定（active、subject、audience、expiresAt），附 `ToConnectorAuthorizationDomain()` |
| `internal/domain/models/domains/connector_authorization_domain.go` | **Add** | `IsGrantedTo(protectedResourceUrl)`（active 且 audience 去結尾斜線後相等）、`RememberedUntil(now)`（min(exp, now+60s)；不通過者 now+60s）、`ToDto()` |
| `internal/domain/models/domains/connector_authorization_errors.go` | **Add** | `ErrConnectorAuthorizationRejected`、`ErrReconnectRequired`（「請到 Claude Code 的 /mcp 選單重新連線」） |
| `internal/domain/models/dto/connector_authorization_dto.go` | **Add** | 通過確認的授權：`Subject`、`ExpiresAt` |
| `internal/domain/service/connector_authorization_service.go` | **Add** | `VerifyConnectorAuthorization`：先查記住的結果，沒有才問交易服務；問不到不記、回 `ErrTradingServiceUnreachable` |
| `internal/application/connector_authorization_application.go` | **Add** | 用例入口 |
| `internal/controller/connector_authorization_controller.go` | **Add** | `Guard(handler)` 包 `auth.RequireBearerToken`（被拒 → `auth.ErrInvalidToken` → 401 + `WWW-Authenticate: Bearer resource_metadata=...`；問不到交易服務 → 非 401 錯誤）、`MetadataHandler()` 包 `auth.ProtectedResourceMetadataHandler`；TokenInfo 帶 `UserID=subject`，SDK 因此擋下跨使用者搶用連線 |
| `internal/infrastructure/persistence/connector_authorization_verdict_repository.go` | **Add** | 記憶體實作，鍵為 token 的 SHA-256；存入時順手清掉過期項，不會只長不縮 |
| `internal/infrastructure/tradingservice/trading_service_proxy.go` + `connector_authorization_inspection_wire.go` | **Modify / Add** | 以 form `token=...` POST `/oauth/introspection`；非 2xx 或看不懂 → `ErrTradingServiceUnreachable`；刪 `session_tokens_wire.go` |
| `internal/domain/service/api_tool_service.go` | **Modify** | 不再依賴 `AuthenticationService`：一律帶 `ToolCallDto.AccessToken`；交易服務回「不認得」→ `ToolOutcomeReconnectRequired` + `ErrReconnectRequired`，不重試 |
| `ApiToolDomain` / `ToolDefinitionDto` / `TradingServiceRequestVo` | **Modify** | 拿掉 `requiresSignIn` / `RequiresSignIn` / `CarriesIdentity`：整個端點都已要求授權，這一格不再有任何作用（死碼） |
| `ToolCallDto`、`mcpCaller`、`ApiToolController` | **Modify** | 拿掉 `SessionKey`；`SuppliedAccessToken` → `AccessToken`；刪掉每件能力的「請先登入」附註 |
| `dto.ToolOutcome`、`FailureReasonDomain` | **Modify** | 刪 `signInRequired` / `signInExpired`；加 `reconnectRequired` |
| 身分保管一整組 | **Remove** | `AuthenticationController/Application/Service`、`session_renewal_gate.go`、`ISignedInSessionRepository`（含 mock 與實作、測試）、`SignedInSessionDomain`、`sign_in_errors.go`、`SignInDto`、`SignInOutcomeDto`、`SignedInSessionDto`、`SessionGrantVo`、`SessionKeyVo`、`TokenPairVo` |
| `cmd/server/tool_catalog.go` | **Modify** | 刪 `trading_register_user`；「我是誰」說明改成確認這份外掛授權代表誰 |
| `cmd/server/config.go` | **Modify** | 加 `PUBLIC_BASE_URL`（預設 `http://localhost:8090`）、`TRADING_SERVICE_PUBLIC_URL`（預設 `http://localhost:8080`）；`ProtectedResourceUrl()`、`ResourceMetadataUrl()` 由設定導出（不從請求推算） |
| `cmd/server/dependencies.go` | **Modify** | 組裝新服務；`/.well-known/oauth-protected-resource` 與 `/.well-known/oauth-protected-resource{MCP_PATH}` 掛 metadata；MCP 路徑包 `Guard`；Instructions 改寫 |
| `IDLE_CONNECTION_TIMEOUT_MINUTES` | **Keep** | SDK 的連線（session）仍然存在、仍需閒置回收；只改說明（放掉後 Claude Code 自己重開，不必重新授權） |
| `/health` | **Not touched** | 外掛自己的存活檢查不經授權（k8s probe 用），不是能力 |
| `IClock` | **Not touched** | 記住多久需要可控的時間 |
| 能力清單其餘宣告的路徑、欄位、說明 | **Not touched** | 只刪 `requiresSignIn` 那一個引數 |

## 3. New Classes / Modules

| Unit | Layer | Responsibility | PRD scenarios |
| :--- | :--- | :--- | :--- |
| `ConnectorAuthorizationInspectionVo` | domain/vo | 交易服務原話的判定 | US-01 |
| `ConnectorAuthorizationDomain` | domain/domains | 判定算不算數、記多久 | US-01 全部、US-02 |
| `ConnectorAuthorizationDto` | domain/dto | 通過者的主體與到期 | US-01 |
| `IConnectorAuthorizationVerdictRepository` / `ConnectorAuthorizationVerdictRepository` | interface / persistence | 以指紋記住判定到期為止 | US-02 |
| `ConnectorAuthorizationService` | domain/service | 查記住的 → 問交易服務 → 判定 → 記住 | US-01、US-02 |
| `ConnectorAuthorizationApplication` | application | 用例入口 | US-01 |
| `ConnectorAuthorizationController` | controller | 把判定接到 SDK 的 bearer 中介層與授權說明 | US-01、US-03 |

## 4. Modified Components

| Component | Current role | Change needed |
| :--- | :--- | :--- |
| `ApiToolService` | 決定用誰的身分、過期自己換新重試 | 一律用呼叫端帶來的授權；不認得 → 請重新連線 |
| `TradingServiceProxy` | 登入 / 換新 / 撤銷 + 代辦 | 確認授權 + 代辦 |
| `buildHttpHandler` | 只掛 MCP 與 health | 加兩條授權說明、MCP 包 Guard |

## 5. Component Relationships

```
HTTP ─▶ /.well-known/oauth-protected-resource[MCP_PATH] ─▶ ProtectedResourceMetadataHandler
HTTP ─▶ MCP_PATH ─▶ RequireBearerToken(verifier) ─▶ StreamableHTTPHandler ─▶ ApiToolController
                          │
                          ▼
          ConnectorAuthorizationApplication ─▶ ConnectorAuthorizationService
                 ├─ IConnectorAuthorizationVerdictRepository（sha256 → 判定, 到期時刻）
                 └─ ITradingServiceProxy.InspectConnectorAuthorization ─▶ POST /oauth/introspection
                        └─ vo.ToConnectorAuthorizationDomain().IsGrantedTo(resource)

ApiToolController ─▶ ApiToolService ─▶ ITradingServiceProxy.Send(request, 呼叫端 bearer)
                                        └─ 401 → reconnectRequired
```

## 6. Extensibility & Handoff Notes

- **Most likely next requirement:** 細分權限（scope），或讓外掛只接受特定 client。
- **Where it lands:** `ConnectorAuthorizationInspectionVo` 多一格、`ConnectorAuthorizationDomain.IsGrantedTo` 多一個條件；
  HTTP 那一面 SDK 的 `RequireBearerTokenOptions.Scopes` 已現成。
- 記住多久只有 `ConnectorAuthorizationDomain` 一處常數。

## 7. Traceability

| PRD Scenario | Fulfilled by |
| :--- | :--- |
| US-01 有效授權照常接待 | Controller Guard + Service + Domain `IsGrantedTo` |
| US-01 沒帶授權不接待、指出說明 | SDK `RequireBearerToken` + `ResourceMetadataURL` |
| US-01 失效授權不接待 | Domain（inactive）→ `ErrConnectorAuthorizationRejected` → `auth.ErrInvalidToken` |
| US-01 別的服務的授權不算數 | Domain `IsGrantedTo` |
| US-01 只差結尾斜線算同一個 | Domain `IsGrantedTo` 去結尾斜線 |
| US-01 問不到交易服務不說成授權失效 | Service 回 `ErrTradingServiceUnreachable` → Controller 非 401 |
| US-02 一分鐘內不再確認／滿一分鐘重問／不超過到期／問不到不記 | Domain `RememberedUntil` + Repository + Service |
| US-03 授權說明、兩個位址同一份 | Controller `MetadataHandler` + `buildHttpHandler` |
| US-04 帶著授權去問（改與看都是） | `ApiToolService` 一律帶 `AccessToken` |
| US-04 不認得 → 重新連線、不重試 | `ApiToolService` + `ErrReconnectRequired` |
| US-04 尚未開通原話帶回 | 既有 refused 原話轉述 |
| US-05 清單與說明 | `tool_catalog.go`、刪 `AuthenticationController`、`dependencies.go` Instructions、`ApiToolController.toMcpTool` |

## 8. Risks & Open Decisions

- 部署順序：交易服務授權伺服器與部署設定（`PUBLIC_BASE_URL`、`TRADING_SERVICE_PUBLIC_URL`）必須先到位。
- 交易服務的 introspection 與登入共用限流器，所有使用者的確認都從外掛同一個來源發出；一分鐘的記住大幅降低呼叫量，但使用者多時仍可能被限流（屆時回「連不到交易服務」，不會誤導成重新授權）。
- 問不到交易服務時 SDK 回 500 而不是 503——SDK 只分 401/400/500；重點是不回 401，Claude Code 不會因此要求重新授權。
- 決定（代替提問）：每一件能力都轉送授權；不再區分需不需要身分；連線（session）保留為有狀態，閒置回收設定保留。
