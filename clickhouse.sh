#!/bin/bash

docker run -d \
    --cap-add=SYS_NICE --cap-add=NET_ADMIN --cap-add=IPC_LOCK \
    --name tgbase-clickhouse \
    --ulimit nofile=262144:262144 \
    -v ./schema.sql:/docker-entrypoint-initdb.d/schema.sql \
    -p 18123:8123 -p19000:9000 \
    clickhouse/clickhouse-server
