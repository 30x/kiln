FROM scratch

COPY build/shipyard /

CMD ["/shipyard"]