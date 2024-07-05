FROM golang:1.22-alpine AS build_stage

RUN apk --no-cache add ca-certificates
WORKDIR /go/src/github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/email-service
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .
ENV GOCACHE=/root/.cache/go-build
RUN --mount=type=cache,target="/root/.cache/go-build" CGO_ENABLED=0 GOOS=linux go build -o /email-service ./cmd/main.go


FROM scratch AS email_stage
COPY --from=build_stage /email-service /email-service
ENTRYPOINT ["/email-service"]
