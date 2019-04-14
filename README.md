# sim_http_server
A golang port of `python -m SimpleHTTPServer`

Principle:

1 With logging for monitor incomer.

2 Use golang's official packets as we can.


```shell
$ ./sim_http_server
2019/04/14 15:48:26 Serving HTTP on 0.0.0.0:8100
192.168.0.102 - - [14/Apr/2019:15:48:28 +0800] "GET / HTTP/1.1" 200 263
192.168.0.102 - - [14/Apr/2019:15:48:30 +0800] "GET / HTTP/1.1" 200 263
^C
```