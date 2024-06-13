FROM grokzen/redis-cluster:7.0.10
ENV REDIS_CLUSTER_IP 0.0.0.0
HEALTHCHECK --interval=30s --timeout=10s --retries=3 \
    CMD redis-cli -c -p 7000 cluster info || exit 1