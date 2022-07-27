FROM golang:1.18

COPY migrations /migrations/

RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

COPY start.sh /start.sh
RUN chmod +x /start.sh

ENTRYPOINT ["/start.sh"]

ENV PG_AUTH_USER "als_auth"
ENV PG_DB "als_db"