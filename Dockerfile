FROM golang:1.23

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o chat-server cmd/server/main.go

CMD ["/app/chat-server"]