FROM golang:1.18-alpine

WORKDIR /app
COPY go.mod ./
COPY go.sum ./
RUN go mod download
RUN go mod verify

COPY . ./
RUN CGO_ENABLED=0 go build -o ./build/goapp ./cmd
EXPOSE 9999
CMD ["./build/goapp"]