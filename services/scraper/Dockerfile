FROM golang:1.21 as BUILD

WORKDIR /app

COPY ./ ./
RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -o scrape

FROM golang:1.21

WORKDIR /app
COPY --from=BUILD /app/scrape /app/scrape

CMD ["/app/scrape"]