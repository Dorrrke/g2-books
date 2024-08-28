FROM golang:1.22.4

WORKDIR /app

COPY . .

RUN go build cmd/g2-books/main.go

CMD ["./main"]