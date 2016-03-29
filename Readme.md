$ make
$ ./bin/kitsession &
[1] 30937

$ ts=2016-03-29T06:11:06Z info=kitsession

$ curl http://localhost:6502/id/1
level=debug ts=2016-03-29T06:11:14Z message="have session id" session_id=1
level=debug ts=2016-03-29T06:11:14Z message="init session" session_id=1
::1 - - [28/Mar/2016:23:11:14 -0700] "GET /id/1 HTTP/1.1" 200 16 "" "curl/7.30.0"
{"Status":"OK"}

$ curl http://localhost:6502/id/1
level=debug ts=2016-03-29T06:11:14Z message="have session id" session_id=1
level=debug ts=2016-03-29T06:11:14Z message="have session context" session_id=1
level=debug ts=2016-03-29T06:11:14Z message="have session expiry" session_id=1 expires=2016-03-29T06:11:19Z
level=debug ts=2016-03-29T06:11:14Z message="touching session" session_id=1
::1 - - [28/Mar/2016:23:11:14 -0700] "GET /id/1 HTTP/1.1" 200 16 "" "curl/7.30.0"
{"Status":"OK"}

$ sleep 5

$ curl http://localhost:6502/id/1
level=debug ts=2016-03-29T06:11:24Z message="have session id" session_id=1
level=debug ts=2016-03-29T06:11:24Z message="have session context" session_id=1
level=debug ts=2016-03-29T06:11:24Z message="have session expiry" session_id=1 expires=2016-03-29T06:11:19Z
level=debug ts=2016-03-29T06:11:24Z message="session expired" session_id=1 expires=2016-03-29T06:11:19Z
level=debug ts=2016-03-29T06:11:24Z message="init session" session_id=1
::1 - - [28/Mar/2016:23:11:24 -0700] "GET /id/1 HTTP/1.1" 200 16 "" "curl/7.30.0"
{"Status":"OK"}
