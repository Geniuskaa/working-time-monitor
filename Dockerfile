FROM golang:1.18 AS build

WORKDIR /app
COPY go.mod go.sum ./

RUN go mod download
RUN go mod tidy

COPY . ./
RUN CGO_ENABLED=0 go build -o ./build/goapp ./cmd

FROM alpine

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /app/configs /configs
COPY --from=build /app/build/goapp /goapp
EXPOSE 7001
CMD ["./goapp"]

