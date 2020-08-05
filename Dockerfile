FROM alpine:3.11
WORKDIR /
COPY  /bin/manager .

ENTRYPOINT ["/manager"]
