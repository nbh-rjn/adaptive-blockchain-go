FROM golang:1.23

WORKDIR /app

COPY . .

RUN go build -o blockchain-system .

CMD ["./blockchain-system"]
