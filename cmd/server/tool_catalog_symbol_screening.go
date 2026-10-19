package main

import (
	"time"

	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/domains"
	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/vo"
)

// administratorOnlyNote is attached to every screening ability, since a refusal for not being an administrator must never read as something to fix and resend.
const administratorOnlyNote = "\n\n**只有管理員能用。** 不是管理員時交易服務會拒絕並說「這是管理員功能」——**照原話轉達、不要重試**，" +
	"也不要教使用者怎麼變成管理員：管理員只由營運者在系統外開通，這個外掛與交易服務都沒有開通的方法。" +
	"想先確認時可以用 trading_get_current_user 看 isAdmin。"

const symbolScreeningVerdictNote = "\n\n**怎麼讀結果**：candidates 已依推薦順序排好（recommendationRank 1 是最值得考慮的）。每個候選標的的 status 四種結論——" +
	"passed（通過：基本條件與策略移植都過了）、rejected（淘汰：stage 說卡在哪一關——basicCondition 是幣安沒有／停止交易／歷史太短／成交太冷清，" +
	"portability 是交易太少／多數區段虧損／去掉最好的一段就虧損——reason 是原因，照原話說）、unfinished（無法完成：行情取不到、策略腳本出錯等，reason 是原因，" +
	"不是這個標的不好）、cancelled（篩選被取消前還沒輪到）；還在跑時另有 waiting／processing。" +
	"\n\nfindings 是判斷用的數字：historyDays、averageDailyQuoteVolume、segments（各區段的 returnRate、maximumDrawdown、closedTradeCount、liquidationCount）、" +
	"returnWithoutBestSegment（拿掉最好的一段後其餘區段連乘的報酬）。通過的標的在設定有現有組合時另有 contribution：" +
	"**gapReductionDays 是加入這個標的後，現有組合沒有任何部位的最長空窗縮短了幾天**（baselineLongestGapDays → combinedLongestGapDays），" +
	"inMarketShare 是它自己有部位的時間比例、overlapShare 是它有部位時現有組合也有部位的比例。contributionFailureReason 有字時代表現有組合比不出來，通過與否不受影響。" +
	"\n\n**通過只代表這套策略原封不動在它身上經得起這次的分段檢查，不是買賣建議**——說結果時要把這句話帶到。"

// symbolScreeningProfileWriteParameters are shared by creating and rewriting a profile, since the trading service reads both the same way.
func symbolScreeningProfileWriteParameters() []vo.ToolParameterVo {
	return []vo.ToolParameterVo{
		bodyParameter("name", vo.ToolParameterKindString, "這份篩選設定叫什麼；同一位管理員的設定不可同名", true),
		bodyParameter("strategyScriptId", vo.ToolParameterKindInteger,
			"要用哪一支策略腳本：必須是**自己的**、吃合約行情（contractKCandle）、產出 signal、"+
				"**不宣告訂單流價位寬度**（篩選用的行情沒有逐筆成交明細）", true),
		bodyParameter("aggregationInterval", vo.ToolParameterKindString,
			"看盤間隔，**只能是 4h 或 1d**", true),
		bodyParameter("parameterValues", vo.ToolParameterKindArray,
			"這支腳本的旋鈕要調成多少，每個為 {\"name\":…, \"value\":…}；沒給的照腳本的預設值", false),
		bodyParameter("tradingMode", vo.ToolParameterKindString,
			"longShort／longOnly／shortOnly，不給即 longOnly", false),
		bodyParameter("leverage", vo.ToolParameterKindString, "槓桿倍數（字串形式的精確小數），不給即 1", false),
		bodyParameter("entryCostPercentage", vo.ToolParameterKindString,
			"每邊手續費百分比（0.05 就是 0.05%），不給即 0.05", false),
		bodyParameter("slippagePercentage", vo.ToolParameterKindString, "滑價百分比，不給即 0.05", false),
		bodyParameter("historyDays", vo.ToolParameterKindInteger,
			"要看多長的歷史（天，180～1825），不給即 730；也是「歷史至少要多長」的門檻", false),
		bodyParameter("segmentCount", vo.ToolParameterKindInteger, "把歷史平均切成幾段各自重演（2～6），不給即 3", false),
		bodyParameter("minimumTradeCount", vo.ToolParameterKindInteger,
			"全部區段平倉筆數合計至少要幾筆才判斷得出來（1～200），不給即 6", false),
		bodyParameter("minimumAverageDailyQuoteVolume", vo.ToolParameterKindString,
			"最近 30 天日均成交額至少要多少 USDT（字串形式的精確小數），不給即 20000000", false),
		bodyParameter("baselineSymbols", vo.ToolParameterKindArray,
			"現有組合標的（例如 [\"BTCUSDT\"]，最多 10 個），用來算通過的標的能把空窗縮短多少；不給就不比較", false),
	}
}

func symbolScreeningApiTools(replayWaitLimit time.Duration) []domains.ApiToolDomain {
	return []domains.ApiToolDomain{
		domains.NewApiToolDomain(
			"trading_list_symbol_screening_profiles",
			"列出自己的篩選設定：每一份記著用哪支策略腳本、旋鈕值、看盤間隔、交易模式、成本、歷史天數、區段數、門檻與現有組合標的。"+
				administratorOnlyNote,
			vo.RequestVerbRead, "/symbol-screening-profiles",
		),
		domains.NewApiToolDomain(
			"trading_get_symbol_screening_profile",
			"看一份自己的篩選設定。別人的或不存在的回「找不到」。"+administratorOnlyNote,
			vo.RequestVerbRead, "/symbol-screening-profiles/{id}",
			pathParameter("id", "篩選設定 id"),
		),
		domains.NewApiToolDomain(
			"trading_create_symbol_screening_profile",
			"建立一份篩選設定：一份命名的篩法——用哪支策略、什麼條件——之後送出篩選時指名它。"+
				"\n\n**只送使用者說了的項目**，沒說的交給交易服務用預設值（歷史 730 天、切 3 段、最少 6 筆、日均成交額 20,000,000 USDT、1 倍、手續費與滑價各 0.05%、只做多），"+
				"不要替使用者把預設值填進去。策略腳本、看盤間隔或數字不合規定時交易服務會拒絕並說原因，照原話帶回。"+
				administratorOnlyNote,
			vo.RequestVerbSubmit, "/symbol-screening-profiles",
			symbolScreeningProfileWriteParameters()...,
		),
		domains.NewApiToolDomain(
			"trading_update_symbol_screening_profile",
			"改寫一份自己的篩選設定。**這是整份改寫**：沒給的項目會回到預設值，不是保留原本的——要保留的請照抄原本的值。"+
				"已經送出的篩選照送出當下的內容跑，不受這次修改影響。"+administratorOnlyNote,
			vo.RequestVerbReplace, "/symbol-screening-profiles/{id}",
			append([]vo.ToolParameterVo{pathParameter("id", "要改寫哪一份篩選設定")},
				symbolScreeningProfileWriteParameters()...)...,
		),
		domains.NewApiToolDomain(
			"trading_delete_symbol_screening_profile",
			"刪除一份自己的篩選設定。有排隊中或進行中的篩選正在用它時會被拒絕，等那筆結束或取消再刪；"+
				"已經結束的篩選仍看得到當時的設定內容。"+administratorOnlyNote,
			vo.RequestVerbRemove, "/symbol-screening-profiles/{id}",
			pathParameter("id", "要刪除哪一份篩選設定"),
		),
		domains.NewApiToolDomain(
			"trading_submit_symbol_screening_run",
			"送出一次標的篩選：指名一份篩選設定，加一串用逗號分隔的永續合約標的，系統在背後逐一檢查"+
				"「這套策略原封不動能不能移植過去、能不能錯開現有部位」。"+
				"\n\n**symbols 照使用者貼的原樣放進去**——全形逗號、大小寫、重複由交易服務整理，外掛不要先整理；一次最多幾個由交易服務決定（預設 20），"+
				"超過或含字母數字以外的字元會被拒絕並說明原因，照原話帶回。"+
				"\n\n**它不等跑完**：立刻回一筆排隊中的篩選（id、status），**把篩選編號說給使用者**，再用 trading_get_symbol_screening_run 看。"+
				"status 是 queued 代表前面還有在跑的（全系統一次只跑一筆），**會自動開始，不要重送**；"+
				"自己已經有一筆在排隊或進行中時會被拒絕，等它結束或用 trading_cancel_symbol_screening_run 取消再送。"+
				"送出當下會記下設定與策略腳本的內容，之後改設定或腳本不影響這一筆。"+
				administratorOnlyNote,
			vo.RequestVerbSubmit, "/symbol-screening-runs",
			bodyParameter("profileId", vo.ToolParameterKindInteger, "用哪一份自己的篩選設定", true),
			bodyParameter("symbols", vo.ToolParameterKindString,
				"用逗號分隔的永續合約標的，照使用者貼的原樣，例如 \"xrpusdt, DOGEUSDT，linkusdt\"", true),
		),
		domains.NewApiToolDomain(
			"trading_list_symbol_screening_runs",
			"列出自己的標的篩選，新到舊，回覆帶 totalCount；**回覆時說出總數**。每一筆帶 status、progressPercent、"+
				"processedCandidateCount／candidateCount、送出與結束時間、failureReason。要看每個標的的結論再拿 id 用 trading_get_symbol_screening_run。"+
				administratorOnlyNote,
			vo.RequestVerbRead, "/symbol-screening-runs",
			queryParameter("status", vo.ToolParameterKindString,
				"只看這一種狀態：queued／running／succeeded／failed／cancelled。不給即全部", false),
			queryParameter("limit", vo.ToolParameterKindInteger, "最多回幾筆，預設 20、最多 200", false),
		),
		domains.NewApiToolDomain(
			"trading_get_symbol_screening_run",
			"看一筆標的篩選：status 五選一——queued（排隊中，會自動開始）、running（進行中，progressPercent 以處理完幾個標的算）、"+
				"succeeded（完成，就算全部被淘汰也是完成）、failed（failureReason 照原話說：「超過時間上限」是篩選太久、「被重新啟動打斷」是交易服務換了一台，都不是策略壞了）、"+
				"cancelled（已取消，處理完的標的保留結果）。snapshot 是送出當下用的設定。"+
				"\n\n**可以順便等一下**：給 waitSeconds（最多 60），篩選在那之前結束就立刻回，沒結束就回當時的狀態——不必每隔幾秒戳一次。"+
				"不是自己的回「找不到」。"+
				symbolScreeningVerdictNote+
				administratorOnlyNote,
			vo.RequestVerbRead, "/symbol-screening-runs/{id}",
			pathParameter("id", "送出篩選時回的那個 id"),
			queryParameter("waitSeconds", vo.ToolParameterKindInteger,
				"最多等幾秒再回（0 到 60）。篩選在那之前結束就提前回；不給即不等", false),
		).Waiting(replayWaitLimit),
		domains.NewApiToolDomain(
			"trading_cancel_symbol_screening_run",
			"取消一筆還沒跑完的標的篩選：排隊中的永遠不會開始；進行中的停在手上那個標的之後，已處理完的標的保留結果，其餘標為已取消。"+
				"已經結束的不能取消，會被拒絕並說它已經結束。"+administratorOnlyNote,
			vo.RequestVerbSubmit, "/symbol-screening-runs/{id}/cancel",
			pathParameter("id", "要取消哪一筆標的篩選"),
		),
	}
}
