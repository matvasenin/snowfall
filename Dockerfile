# Build the dependency management image
FROM ghcr.io/astral-sh/uv:python3.13-alpine AS builder
# Enable bytecode optimization
ENV UV_COMPILE_BYTECODE=0
# Disable hardlinks
ENV UV_LINK_MODE=copy
# Force UV to use system-wide interpreter
ENV UV_PYTHON_DOWNLOADS=0

# Install dependencies with caching
WORKDIR /app
RUN --mount=type=cache,target=/root/.cache/uv \
    --mount=type=bind,source=uv.lock,target=uv.lock \
    --mount=type=bind,source=pyproject.toml,target=pyproject.toml \
    uv sync --locked --no-install-project --no-dev
COPY . /app
RUN --mount=type=cache,target=/root/.cache/uv \
    uv sync --locked --no-dev

# Build the production image
FROM python:3.13-alpine

# Copy the application from the builder
COPY --from=builder /app /app

# Place executables in the environment at the front of the path
ENV PATH="/app/.venv/bin:$PATH"

# Use `/app` as the working directory
WORKDIR /app

CMD ["python", "main.py"]