FROM golang:1.19.7-alpine3.17 AS build

WORKDIR /opt/

RUN apk add --update pkgconf git build-base make

COPY . .

RUN make staticbuild

FROM alpine

ENV GIN_MODE=release

COPY --from=build /opt/main .

ENTRYPOINT ["./main"]
