FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY server ./server
RUN CGO_ENABLED=0 go build -trimpath -o /out/homeflow-api ./server/cmd/api

FROM alpine:3.22
RUN adduser -D -u 10001 app
USER app
COPY --from=build /out/homeflow-api /homeflow-api
EXPOSE 8080
ENTRYPOINT ["/homeflow-api"]
