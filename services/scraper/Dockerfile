FROM golang:1.18

WORKDIR /app

COPY ./ ./
RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -o scrape

CMD ["./scrape"]