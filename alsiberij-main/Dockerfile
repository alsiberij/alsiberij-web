FROM golang:1.18 AS builder

WORKDIR /app

COPY . .
RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -a -o app .

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/app /bin/app

COPY view /bin/view
ENV PATH_VIEW "/bin/view"

ENV PORT 11400
EXPOSE $PORT

ENTRYPOINT ["/bin/app"]
