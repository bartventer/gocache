FROM memcached:latest
HEALTHCHECK --interval=30s --timeout=10s --retries=3 \
    CMD echo stats | nc localhost 11211 || exit 1