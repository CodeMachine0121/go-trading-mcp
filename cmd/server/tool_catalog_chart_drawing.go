package main

import (
	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/domains"
	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/vo"
)

const chartDrawingAssistantOnlyNote = "\n\n**只能動助手畫的**：這個外掛只能新增與移除助手繪圖（source 是 assistant），" +
	"**使用者手畫的（source 是 manual）動不了**，也沒有任何能力可以移動、改色、鎖定任何一個繪圖——包括助手自己畫的，要改就移除再畫。" +
	"使用者要改或刪他手畫的線時，請他自己在網頁的看盤工作區上動，**不要改用別的辦法去刪**。"

const chartDrawingBrushNote = "\n\n**奇異筆（kind 是 brush）**：使用者按住拖曳畫的自由曲線，只會是手畫的。" +
	"它的 anchors 是沿筆跡的一連串點（2 到 200 個），**openTime 不一定是 K 線的起始時間**（可能落在兩根 K 線之間），" +
	"不要把它的點當成 K 線起始時間拿去畫線。"

const chartDrawingSeeItOnTheWebNote = "\n\n**畫完要告訴使用者**：回到網頁的看盤工作區（或重新聚焦那一頁）才看得到新畫的線，畫面不會自己跳出來。"

func chartDrawingBoardParameters() []vo.ToolParameterVo {
	return []vo.ToolParameterVo{
		pathParameter("market",
			"哪一個市場的繪圖板，二選一：spot（現貨）或 contract（永續合約）。同一個代號在現貨與合約各是一份繪圖板，互不相干"),
		pathParameter("symbol", "交易標的，如 BTCUSDT（大小寫不拘）"),
	}
}

func chartDrawingApiTools() []domains.ApiToolDomain {
	return []domains.ApiToolDomain{
		domains.NewApiToolDomain(
			"trading_get_chart_drawings",
			"看使用者在一個市場、一個交易標的上的**繪圖板**：回覆 market、symbol、revision（版本，每改一次前進一格）與 drawings。"+
				"每一個繪圖帶 id（編號）、kind、anchors（openTime 是那一根 K 線的起始時間——奇異筆例外，見下方；price 是十進位字串）、color、lineWidth、lineStyle、locked、visible，"+
				"以及 source：manual（使用者手畫）或 assistant（助手畫的）。**從沒畫過的繪圖板回 revision 0、drawings 是空的**，不是錯誤。"+
				"\n\n**畫線之前先讀一次**，別重複畫已經有的線；要移除助手畫的線時，從這裡找出 source 是 assistant 的編號。"+
				chartDrawingBrushNote+
				chartDrawingAssistantOnlyNote,
			vo.RequestVerbRead, "/users/me/chart-drawings/{market}/{symbol}",
			chartDrawingBoardParameters()...,
		),
		domains.NewApiToolDomain(
			"trading_add_chart_drawings",
			"在使用者的 K 線圖上畫線：在一份繪圖板上新增一批**助手繪圖**（1 到 20 個），整批檢查、整批生效。"+
				"回覆帶 revision（新版本）與 addedDrawings（新畫的每一個，含編號與吸附後的實際價格）。"+
				"繪圖不涉及錢，**新增直接生效，不必先問使用者確認**。"+
				"\n\n**先讀行情，不要自己編時間或價位**：先用彙總 K 線序列找到真正的 K 線起始時間與高低點——"+
				"現貨板（spot）用 trading_get_k_candle_series，合約板（contract）用 trading_get_contract_k_candle_series，"+
				"**interval 給與這次相同的刻度**（不要用 displayableCandleCount 讓系統挑別的刻度；一分鐘原始 K 線只對得上 1m），"+
				"回來每一根的 openTime 與開高低收就是能用的錨點。**每一個錨點的 openTime 必須正好是那個刻度上一根已經存在的 K 線的起始時間**"+
				"（一小時刻度 08:30 不是起始時間、還沒走到的未來沒有 K 線、台股週末沒有 K 線）。"+
				"\n\n**價格優先用吸附**：錨點給 snapTo（open／high／low／close）就由交易服務取那一根的開高低收（現貨板取現貨 K 線、合約板取合約 K 線），"+
				"免得畫在編造的價位；真的要畫在不是開高低收的價位才給 price。**price 與 snapTo 一定要給其中一個、不能兩個都給**。"+
				"\n\n**奇異筆（brush）不能畫**：奇異筆只能由使用者手畫，助手不能新增；送了交易服務會整批拒絕。"+
				"使用者要你用奇異筆圈出什麼時，改用可畫的種類（例如 rectangle 矩形），或請他自己在網頁上畫。"+
				"\n\n**上限**：一份繪圖板最多 200 個、一次最多 20 個；超過整批拒絕，拒絕會說還剩幾個空位。"+
				"\n\n被拒時拒絕會指出第幾個繪圖第幾個錨點、為什麼——**照它說的改一改再送，不要原樣重送**；整批被拒時一個都沒畫上去。"+
				chartDrawingAssistantOnlyNote+
				chartDrawingSeeItOnTheWebNote,
			vo.RequestVerbSubmit, "/users/me/chart-drawings/{market}/{symbol}/assistant-drawings",
			append(chartDrawingBoardParameters(),
				bodyParameter("interval", vo.ToolParameterKindString,
					"這批繪圖是看哪一種刻度的 K 線畫的，六選一：1m、5m、15m、1h、4h、1d。每一個錨點的 openTime 都必須是這個刻度上一根 K 線的起始時間", true),
				bodyParameter("drawings", vo.ToolParameterKindArray,
					"要新增的繪圖，1 到 20 個，後面的疊在前面的上面。每一個形狀："+
						"{\"kind\":\"horizontalLine\", \"anchors\":[{\"openTime\":\"2026-10-19T08:00:00Z\", \"snapTo\":\"low\"}], "+
						"\"color\":\"red\", \"lineWidth\":2, \"lineStyle\":\"dashed\"}。"+
						"\n\nkind 與它要的錨點數：trendLine（趨勢線）、ray（射線）、fibonacciRetracement（斐波那契回撤）、rectangle（矩形）、"+
						"dateAndPriceRange（日期與價格範圍）各 **2 個**；horizontalLine（水平線）、horizontalRay（水平射線）、verticalLine（垂直線）各 **1 個**。"+
						"brush（奇異筆）只能由使用者手畫，不能新增。"+
						"\n\n每一個錨點：openTime（RFC3339，那一根 K 線的起始時間）加上 snapTo（open／high／low／close，優先用）或 price（大於零，字串或數字）二選一。"+
						"\n\n樣式都可以省略，省略即照那一種繪圖的預設樣式：color 十二選一——red、orange、yellow、green、teal、cyan、blue、indigo、purple、pink、gray、silver；"+
						"lineWidth 是 1 到 4 的整數；lineStyle 三選一——solid（實線）、dashed（虛線）、dotted（點線）。"+
						"沒有欄位可以說它是手畫的，新增的一律是助手繪圖", true),
			)...,
		),
		domains.NewApiToolDomain(
			"trading_remove_chart_drawings",
			"移除一份繪圖板上**助手畫的**繪圖：給 ids 只移除指名的那幾個，**不給 ids 即移除這份繪圖板上助手畫的全部**（手畫的留著）。"+
				"回覆帶 revision、removedCount（移除了幾個）與 removedDrawingIds。沒有助手繪圖時移除全部，照常回「移除了 0 個」，不是錯誤。"+
				"\n\n**指名的裡面只要有一個是手畫的，整次拒絕、一個都不刪**；指名一個這份繪圖板上沒有的編號也整次拒絕——不知道編號時先用 trading_get_chart_drawings 找 source 是 assistant 的。"+
				"移除直接生效，不必先問使用者確認，但只移除使用者要你清掉的那些。"+
				chartDrawingAssistantOnlyNote,
			vo.RequestVerbRemove, "/users/me/chart-drawings/{market}/{symbol}/assistant-drawings",
			append(chartDrawingBoardParameters(),
				queryParameter("ids", vo.ToolParameterKindString,
					"要移除的助手繪圖編號，用逗號分隔的一串文字，例如 \"12,15\"（不是陣列）。不給即移除助手畫的全部", false),
			)...,
		),
	}
}
