FROM golang:1.15-alpine AS build

WORKDIR /src/

COPY main.go go.* /src/
ADD . /src/

RUN CGO_ENABLED=0 go build -o /bin/lists-server

FROM scratch
COPY --from=build /bin/lists-server /bin/lists-server
ENTRYPOINT ["/bin/lists-server"]
