FROM golang:1.21-alpine

RUN apk --no-cache add ca-certificates

WORKDIR /go/src/github.com/Hukyl/genesis-kma-school-entry

COPY go.mod ./
RUN go mod download && go mod verify

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -o /api-server .

EXPOSE 8080

ENTRYPOINT ["/api-server"]