FROM ubuntu@sha256:562111b88caa79c71391c0c4d3c2114a7d5c35915c20561fd893cb1e07929e6b
ENV DEBIAN_FRONTEND=noninteractive
RUN apt-get update && apt-get install -y --no-install-recommends build-essential git pkg-config libgtk-4-dev libwebkitgtk-6.0-dev patchelf dpkg-dev xz-utils xvfb xauth dbus-x11 xdotool gstreamer1.0-plugins-base gstreamer1.0-plugins-good fonts-dejavu-core curl ca-certificates python3 file && rm -rf /var/lib/apt/lists/*
ENV GOROOT=/opt/go PATH=/opt/go/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin GOCACHE=/work/.cache/container-go GOMODCACHE=/work/.cache/container-modules TMPDIR=/work/.cache/tmp
WORKDIR /work
