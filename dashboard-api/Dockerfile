# ---- Stage 1: Dependencies -----------------------------------------
FROM python:3.12-slim AS dependencies

WORKDIR /app

RUN pip install --upgrade pip --no-cache-dir
COPY requirements.txt .
RUN pip install --no-cache-dir --user -r requirements.txt

# ---- Stage 2: Runtime ----------------------------------------------
FROM python:3.12-slim AS runtime

ENV PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1 \
    PATH="/home/appuser/.local/bin:$PATH"

RUN addgroup --gid 1001 appgroup \
 && adduser --uid 1001 --gid 1001 --disabled-password --gecos "" appuser

WORKDIR /app

COPY --from=dependencies /root/.local /home/appuser/.local
COPY --chown=appuser:appgroup . .

USER appuser

EXPOSE 8000

CMD ["uvicorn", "app.main:app", "--host", "0.0.0.0", "--port", "8000"]
