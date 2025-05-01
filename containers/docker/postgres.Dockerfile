FROM postgres:16.8-alpine3.21

HEALTHCHECK --interval=5s --timeout=5s --retries=3 CMD pg_isready -U postgres || exit 1