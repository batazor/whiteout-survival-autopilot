FROM ghcr.io/astral-sh/uv:python3.11-bookworm-slim

WORKDIR /app

# 1) Install build deps & OpenCV runtime libs (cacheable layer)
RUN apt-get update \
  && apt-get install -y --no-install-recommends \
       build-essential \
       libgl1 \
       libglib2.0-0 \
  && rm -rf /var/lib/apt/lists/*

# 2) Copy lockfiles and install dependencies only
COPY pyproject.toml uv.lock ./
RUN --mount=type=cache,target=/root/.cache/uv \
    uv sync --locked --no-install-project

# 3) Copy your application code & install the project
COPY screenshot_ocr_service.py ./
RUN uv sync --locked

# 4) Expose port & run your service
EXPOSE 8000
CMD ["uv", "run", "screenshot_ocr_service.py", "--host", "0.0.0.0", "--port", "8000"]
