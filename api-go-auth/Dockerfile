FROM golang:1.18 AS builder

WORKDIR /app

COPY . .
RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -a -o app .

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/app /bin/app

COPY config.example.json /config.json

COPY ssl /bin/ssl
ENV SSL_PATH "/bin/ssl"

ENV LOGS_PATH "/logs"

ENV PORT_SEC 11400
ENV PORT_INSEC 10400

EXPOSE $PORT_SEC
EXPOSE $PORT_INSEC

ENTRYPOINT ["/bin/app"]
