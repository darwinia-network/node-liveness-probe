FROM alpine

COPY bin/node-liveness-probe_linux_amd64 /usr/bin/node-liveness-probe

EXPOSE 49944
ENTRYPOINT [ "/usr/bin/node-liveness-probe" ]
