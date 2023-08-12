FROM golang:1.18 as build

WORKDIR /app

COPY ./ ./
RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -o scrape

FROM scratch

COPY --from=build /app/scrape /app/scrape 

CMD ["./app/scrape"]