FROM golang:1.22-alpine as build_stage

RUN apk --no-cache add ca-certificates
WORKDIR /go/src/github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/currency-rate
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .
ENV GOCACHE=/root/.cache/go-build
RUN --mount=type=cache,target="/root/.cache/go-build" CGO_ENABLED=0 GOOS=linux go build -o /api-server currency-rate/cmd/main.go
RUN --mount=type=cache,target="/root/.cache/go-build" CGO_ENABLED=0 GOOS=linux go build -o /email-service email-service/cmd/main.go

FROM scratch as api_stage
COPY --from=build_stage /api-server /api-server
EXPOSE 8080
ENTRYPOINT ["/api-server"]

FROM scratch as email_stage
COPY --from=build_stage /email-service /email-service
ENTRYPOINT ["/email-service"]
