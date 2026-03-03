FROM golang:1.26.0-alpine3.23 AS builder

WORKDIR /usr/src/app/

ARG SERVICE_NAME

COPY ${SERVICE_NAME}/go.mod ${SERVICE_NAME}/go.sum ./${SERVICE_NAME}/

WORKDIR /usr/src/app/${SERVICE_NAME}
RUN go mod download

COPY ${SERVICE_NAME}/ ./
RUN CGO_ENABLED=0 go build -o /usr/src/app/bin/${SERVICE_NAME} .

FROM scratch

ARG SERVICE_NAME

COPY --from=alpine:latest /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/src/app/bin/${SERVICE_NAME} /app/service
CMD ["/app/service"]