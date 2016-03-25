FROM scratch
# FROM alpine:3.1

COPY build/shipyard /

CMD ["/shipyard"]