from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware

from app.routers import flags

app = FastAPI(
    title="Dashboard API",
    description="Feature Flags Management Control Plane",
    version="1.0.0",
)

app.add_middleware(
    CORSMiddleware,
    allow_origins=["http://localhost:4200"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

app.include_router(flags.router, prefix="/api/v1")


@app.get("/health", tags=["health"])
async def health_check() -> dict:
    return {"status": "ok"}
