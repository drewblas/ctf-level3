curl http://127.0.0.1:9090/healthcheck

curl http://127.0.0.1:9090/index\?path\=test/data/input

curl http://127.0.0.1:9090/isIndexed

curl -w %{time_total}\\n http://127.0.0.1:9090/?q=part