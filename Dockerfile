# FROM scratch
FROM alpine:3.1

RUN apk update
RUN apk add ca-certificates
RUN update-ca-certificates

COPY build/shipyard /

CMD ["/shipyard"]