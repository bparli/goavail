FROM gliderlabs/alpine:3.4

# Necessary depedencies
RUN apk --update add bash curl tar

# Install S6 from static bins
RUN cd / && curl -L https://github.com/just-containers/skaware/releases/download/v1.17.1/s6-eeb0f9098450dbe470fc9b60627d15df62b04239-linux-amd64-bin.tar.gz | tar -xvzf -

COPY goavail /bin/goavail

COPY s6 /etc

RUN chmod +x /etc/services/goavail.svc/run

RUN chmod +x /etc/services/.s6-svscan/crash

RUN chmod +x /etc/services/.s6-svscan/finish

EXPOSE 7946
EXPOSE 8080

CMD ["/bin/s6-svscan", "/etc/services"]
