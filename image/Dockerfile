FROM alpine:latest

RUN apk add --no-cache iperf3 bash bind-tools

COPY debugtool_entrypoint.sh /usr/bin/debugtool_entrypoint.sh

ENTRYPOINT ["bash","-x","/usr/bin/debugtool_entrypoint.sh"]
