FROM busybox:ubuntu

ADD bin/stdiscosrv /stdiscosrv

ENTRYPOINT ["/stdiscosrv", "-http", "-listen", ":8080", "-debug", "-db-backend=postgres", "-db-dsn=postgres://stdiscosrv:kossanmu7@postgres/stdiscosrv?sslmode=disable"]

