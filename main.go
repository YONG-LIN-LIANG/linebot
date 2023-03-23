package main
import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	// 使用go-linebot library
	"github.com/line/line-bot-sdk-go/v7/linebot"
)
var bot *linebot.Client

func main() {
	var err error
	bot, err := linebot.New(
		// 取得環境變數PORT並賦值給port
		os.Getenv("ChannelSecret"), 
		os.Getenv("ChannelAccessToken"))
	log.Println("Bot:", bot, " err:", err)
	// 定義一個路由為/callback，callbackHandler處理這個API請求及回應
	http.HandleFunc("/callback", callbackHandler)
	// 取得環境變數PORT並賦值給port
	port := os.Getenv("PORT")
	// 這邊做addr格式化的事情
	addr := fmt.Sprintf(":%s", port)
	// 設定這個server在哪個port號服務，addr要長成 :port 的格式
	http.ListenAndServe(addr, nil)
}

// callback API Func，帶入response，request
func callbackHandler(w http.ResponseWriter, r *http.Request) {
	// bot.ParseRequest(r) 的用意是將收到的 HTTP 請求轉換成 Line Bot 所需的資料結構，以便後續進行訊息處理及回應。
	// 當 Line Bot 收到使用者的訊息或動作時，Line 伺服器會將相關資訊透過 HTTP POST 請求傳送至開發者指定的 webhook URL，而 bot.ParseRequest(r) 的作用就是將此 HTTP POST 請求轉換成 linebot.Event 類型的資料，以供後續程式進行處理。
	// 在呼叫 bot.ParseRequest(r) 時，我們需要將接收到的 HTTP 請求（http.Request）作為參數傳入，並且需要確保此請求的 body 已經被解析為 JSON 格式，否則會發生解析失敗的錯誤。解析成功後，bot.ParseRequest(r) 會返回一個 []*linebot.Event 的切片，其中每個 linebot.Event 都代表著一個使用者的訊息或動作，開發者可以針對不同的 linebot.EventType 進行相應的處理及回應。
	
	events, err := bot.ParseRequest(r)
	if err != nil {
		if err == linebot.ErrInvalidSignature {
			w.WriteHeader(400)
		} else {
			w.WriteHeader(500)
		}
		return
	}
	for _, event := range events {
		// linebot eventType
		// EventTypeMessage: 處理收到訊息的情況
		// EventTypeFollow: 處理使用者加入好友的情況
		// 
		if event.Type == linebot.EventTypeMessage {
			// event.Message.(type)這種寫法是判斷event.Message的是哪種type，這裡的type 有
			// *linebot.TextMessage: 文字訊息，使用者傳送文字時，就會以這種型別的訊息回應。
			// *linebot.ImageMessage: 圖片訊息，使用者傳送圖片時，就會以這種型別的訊息回應。
			// *linebot.StickerMessage: 貼圖訊息，使用者傳送貼圖時，就會以這種型別的訊息回應。
			// *linebot.VideoMessage: 影片訊息，使用者傳送影片時，就會以這種型別的訊息回應。
			// *linebot.AudioMessage：音訊訊息，使用者傳送音訊時，就會以這種型別的訊息回應。
			// *linebot.LocationMessage：位置訊息，使用者傳送位置時，就會以這種型別的訊息回應。
			// 以下是進階型別:
			// *linebot.TemplateMessage：模板訊息，可以使用預設的樣板設計，包括 ButtonsTemplate、ConfirmTemplate、CarouselTemplate 等，也可以自定義樣板內容。
			// *linebot.FlexMessage：彈性訊息，可以使用彈性框架來設計各種豐富的訊息內容，包括 Bubble、Carousel 等，彈性框架可以使用 JSON 格式設計，也可以使用 Flex Message Builder 來設計。
			// *linebot.ImagemapMessage：圖像地圖訊息，可以讓使用者在圖片上點選不同的區域，並顯示相關訊息內容。
			// *linebot.LocationMessage：位置訊息，除了基本的位置資訊外，還可以設定標題、地址、縮圖等相關訊息內容。
			switch message := event.Message.(type) {
			// 如果message的type是*linebot.TextMessage
			case *linebot.TextMessage:
				// 英文解釋: GetMessageQuota: Get how many remain free tier push message quota you still have this month
				// 中文解釋: 使用 bot.GetMessageQuota().Do() 方法可以讓開發者獲取 LineBot 目前的訊息配額狀態，包括已使用的訊息配額數量、剩餘的訊息配額數量、可使用的訊息配額上限等相關資訊，讓開發者可以掌握 LineBot 的訊息使用情況，以避免超出配額限制而被鎖定。
				quota, err := bot.GetMessageQuota().Do()
				if err != nil {
					log.Println("Quota err:", err)
				}
				// message.ID: Msg unique ID
				// message.Text: Msg text
				// bot.ReplyMessage() 是用來回覆訊息給使用者的方法
				replyToken := event.ReplyToken
				if _, err = bot.ReplyMessage(
					replyToken, 
					linebot.NewTextMessage("msg ID:"+message.ID+":"+"Get:"+message.Text+" , \n OK! remain message:"+strconv.FormatInt(int64(quota.Value), 10))).Do(); 
					err != nil {
						log.Print(err)
					}
			
				// Handle only on Sticker message
			case *linebot.StickerMessage:
				var kw string
				for _, k := range message.Keywords {
					kw = kw + "," + k
				}

				outStickerResult := fmt.Sprintf("收到貼圖訊息: %s, pkg: %s kw: %s  text: %s", message.StickerID, message.PackageID, kw, message.Text)
				if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(outStickerResult)).Do(); err != nil {
					log.Print(err)
				}
			}
		}
	}
}

