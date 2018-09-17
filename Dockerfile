FROM golang:alpine AS build

RUN apk update && apk --no-cache add gcc musl-dev

COPY assets /assets

RUN go build -o /assets/check /assets/check.go && \
    go build -o /assets/in /assets/in.go && \
    go build -o /assets/out /assets/out.go


FROM build AS test

RUN cd /assets/common && \
    go test


FROM alpine:edge AS resource

RUN apk update && apk --no-cache add bash ca-certificates

COPY --from=build /assets/check /opt/resource/check
COPY --from=build /assets/in /opt/resource/in
COPY --from=build /assets/out /opt/resource/out


FROM resource
