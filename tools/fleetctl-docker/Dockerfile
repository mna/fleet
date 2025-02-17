FROM rust:latest AS builder

ARG transporter_url=https://itunesconnect.apple.com/WebObjects/iTunesConnect.woa/ra/resources/download/public/Transporter__Linux/bin

RUN cargo install --version 0.16.0 apple-codesign \
  && curl -sSf $transporter_url -o transporter_install.sh \
  && sh transporter_install.sh --target transporter --accept --noexec

FROM debian:stable-slim

ARG binpath=build/binary-bundle/linux/fleetctl

RUN apt-get update \
  && dpkg --add-architecture i386 \
  && apt update \
  && apt install -y --no-install-recommends ca-certificates cpio libxml2 wine wine32 libgtk-3-0 \
  && rm -rf /var/lib/apt/lists/* 

# copy macOS dependencies
COPY --from=fleetdm/bomutils:latest /usr/bin/mkbom /usr/local/bin/xar /usr/bin/
COPY --from=fleetdm/bomutils:latest /usr/local/lib /usr/local/lib/
COPY --from=builder /transporter/itms /usr/local/
COPY --from=builder /usr/local/cargo/bin/rcodesign /usr/local/bin

# copy Windows dependencies
COPY --from=fleetdm/wix:latest /home/wine /home/wine

# copy fleetctl
COPY ${binpath} /usr/bin/fleetctl

ENV FLEETCTL_NATIVE_TOOLING=1 WINEPREFIX=/home/wine/.wine WINEARCH=win32 PATH="/home/wine/bin:$PATH" WINEDEBUG=-all

ENTRYPOINT ["fleetctl"]
