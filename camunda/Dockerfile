# Camunda Platform 7 Run - lightweight BPM engine
FROM camunda/camunda-bpm-platform:run-latest

# Set environment variables
ENV CAMUNDA_BPM_RUN_AUTH_ENABLED=true
ENV CAMUNDA_BPM_RUN_AUTH_AUTHENTICATION=basic

# Expose port (Railway will use $PORT)
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=60s --retries=3 \
  CMD curl -f http://localhost:8080/engine-rest/engine || exit 1
