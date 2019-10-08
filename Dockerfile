# build stage
FROM golang:1.13

WORKDIR /build

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && apt-get install -y protobuf-compiler \
    && rm -fr /var/lib/apt/lists/*

ENV GO111MODULE=on

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go get -u github.com/golang/protobuf/protoc-gen-go

# Compile all protobuf files for Go in any directory in the source
RUN for D in $(find . -name \*.proto -print0 | xargs -0 -n 1 dirname | sort -u); do (cd $D && protoc -I. --go_out=plugins=grpc:. *.proto); done

RUN CGO_ENABLED=0 GOOS=linux go build -o bin/server server/*.go
RUN CGO_ENABLED=0 GOOS=linux go build -o bin/client client/*.go


# final stage for compiled artifacts
FROM scratch

ENV UNIDB_DIR=/data
ENV UNIDB_LISTEN=0.0.0.0:9002

VOLUME /data

COPY --from=0 /build/bin/server /bin/server
COPY --from=0 /build/bin/client /bin/client

EXPOSE 9002

CMD ["/bin/server"]
