FROM gcr.io/distroless/static-debian11:nonroot
ARG TARGETARCH
COPY bin/gpu-metrics-exporter-$TARGETARCH /usr/local/bin/gpu-metrics-exporter
CMD ["gpu-metrics-exporter"]