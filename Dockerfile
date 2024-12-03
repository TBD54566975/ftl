# Declare build arguments at the very top
ARG RUNTIME=scratch-runtime

# Get certificates from Alpine (smaller than Ubuntu)
FROM alpine:latest AS certs
RUN apk --update add ca-certificates

# Used for everything except ftl-runner
FROM scratch AS scratch-runtime
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Used for ftl-runner
FROM ubuntu:24.04 AS ubuntu-runtime
RUN apt-get update && apt-get install -y ca-certificates
RUN mkdir -p /root/deployments

# Final stage selection
FROM ${RUNTIME}
ARG EXTRA_FILES
ARG SERVICE

WORKDIR /root/
COPY . .


# Common environment variables
ENV PATH="$PATH:/root"

# Service-specific configurations
EXPOSE 8891
EXPOSE 8892
EXPOSE 8893

# Environment variables for all (most) services
ENV FTL_ENDPOINT="http://host.docker.internal:8892"
ENV FTL_BIND=http://0.0.0.0:8892
ENV FTL_ADVERTISE=http://127.0.0.1:8892

# Controller-specific configurations
ENV FTL_CONTROLLER_CONSOLE_URL="*"
ENV FTL_CONTROLLER_DSN="postgres://host.docker.internal/ftl?sslmode=disable&user=postgres&password=secret"

# Provisioner-specific configurations
ENV FTL_PROVISIONER_PLUGIN_CONFIG_FILE="/root/ftl-provisioner-config.toml"

# Default command
CMD ["/root/svc"]
