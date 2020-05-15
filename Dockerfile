FROM geekidea/alpine-a:3.9

RUN apk update \
        && apk upgrade \
        && apk add --no-cache ca-certificates \
        && update-ca-certificates 2>/dev/null || true

ADD ./kubernetes-transform /usr/local/bin/kubernetes-transform
RUN chmod u+x /usr/local/bin/kubernetes-transform

ENTRYPOINT ["kubernetes-transform", "--config",  "/etc/xdhuxc/config.prod.yaml"]
