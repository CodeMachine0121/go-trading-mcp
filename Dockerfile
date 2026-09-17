# 這個外掛沒有資料庫、不寫任何檔案，所以它裝得進一個幾乎什麼都沒有的映像檔。
# 唯一留下的非必要東西是 busybox 的 wget——健康檢查要有人打得了 /health，
# 而「容器活著」與「外掛答得出話」是兩件事。

FROM golang:1.26-alpine AS builder

WORKDIR /src

# 先只複製依賴清單，這樣改程式碼不會讓下載依賴那一層失效。
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# CGO_ENABLED=0：這個外掛一個 C 函式庫都不用（沒有資料庫驅動），
# 關掉它才編得出一個不依賴任何系統函式庫的執行檔。
# -trimpath 拿掉建置機器的路徑，-s -w 拿掉除錯符號——兩者都只是為了讓映像檔小一點。
RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath -ldflags="-s -w" \
    -o /out/server ./cmd/server

FROM alpine:3.21

# 不用 root 跑。這個行程不需要碰任何檔案，所以它連自己的家目錄都不需要。
RUN adduser -D -H -u 10001 connector

COPY --from=builder /out/server /usr/local/bin/server

USER connector

EXPOSE 8090

# 健康檢查打的是外掛自己的 /health，**不是**交易服務的。
# 交易服務掛掉不該讓這個容器被判定為壞掉——它只是暫時沒有東西可以轉達。
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --quiet --spider http://127.0.0.1:8090/health || exit 1

ENTRYPOINT ["/usr/local/bin/server"]
