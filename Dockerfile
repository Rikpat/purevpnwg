FROM alpine:3.19
RUN apk add --no-cache \
      chromium \
      chromium-swiftshader

COPY purevpnwg /bin/purevpnwg

ENV PUREVPN_USERNAME=""
ENV PUREVPN_PASSWORD=""

# Amsterdam
ENV PUREVPN_SERVER_COUNTRY="NL"
ENV PUREVPN_SERVER_CITY="2902"
ENV PUREVPN_WIREGUARD_FILE="/out/wg0.conf"

ENTRYPOINT ["/bin/purevpnwg", "full"]
