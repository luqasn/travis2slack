FROM golang AS builder

COPY . $GOPATH/src/github.com/luqasn/travis2slack
WORKDIR $GOPATH/src/github.com/luqasn/travis2slack/

RUN CGO_ENABLED=0 go build -o /go/bin/travis2slack

FROM scratch

COPY --from=builder /go/bin/travis2slack /go/bin/travis2slack

ENTRYPOINT ["/go/bin/travis2slack"]