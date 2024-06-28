FROM golang:1.22-alpine as build_stage

RUN apk --no-cache add ca-certificates
WORKDIR /go/src/github.com/Hukyl/genesis-kma-school-entry
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .
ENV GOCACHE=/root/.cache/go-build
RUN --mount=type=cache,target="/root/.cache/go-build" CGO_ENABLED=0 GOOS=linux go build -o /api-server .

FROM scratch as production_stage
COPY --from=build_stage /api-server /api-server
EXPOSE 8080
ENTRYPOINT ["/api-server"]
