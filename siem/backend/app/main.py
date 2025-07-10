"""
Main FastAPI application for SIEM Backend API
"""
from fastapi import FastAPI, Request, HTTPException
from fastapi.middleware.cors import CORSMiddleware
from fastapi.middleware.trustedhost import TrustedHostMiddleware
from fastapi.responses import JSONResponse, Response
from contextlib import asynccontextmanager
import structlog
import time
from prometheus_client import Counter, Histogram, generate_latest, CONTENT_TYPE_LATEST

from .config import settings
from .api.v1.api import api_router

# Prometheus metrics
REQUEST_COUNT = Counter('http_requests_total', 'Total HTTP requests', ['method', 'endpoint', 'status'])
REQUEST_LATENCY = Histogram('http_request_duration_seconds', 'HTTP request latency')

logger = structlog.get_logger()


@asynccontextmanager
async def lifespan(app: FastAPI):
    """Application lifespan manager"""
    # Startup
    logger.info("Starting SIEM Backend API", version=settings.app_version)
    
    # Setup logging
    logger.info("Logging configured")
    
    yield
    
    # Shutdown
    logger.info("Shutting down SIEM Backend API")


# Create FastAPI application
app = FastAPI(
    title=settings.app_name,
    version=settings.app_version,
    description="Security Information and Event Management (SIEM) Backend API",
    docs_url="/docs" if settings.debug else None,
    redoc_url="/redoc" if settings.debug else None,
    lifespan=lifespan,
)

# Add CORS middleware
app.add_middleware(
    CORSMiddleware,
    allow_origins=settings.cors_origins,
    allow_credentials=settings.cors_allow_credentials,
    allow_methods=settings.cors_allow_methods,
    allow_headers=settings.cors_allow_headers,
)

# Add trusted host middleware
app.add_middleware(
    TrustedHostMiddleware,
    allowed_hosts=["*"] if settings.debug else ["localhost", "127.0.0.1"]
)


@app.middleware("http")
async def add_process_time_header(request: Request, call_next):
    """Add processing time header and log requests"""
    start_time = time.time()
    
    # Process request
    response = await call_next(request)
    
    # Calculate processing time
    process_time = time.time() - start_time
    
    # Add processing time header
    response.headers["X-Process-Time"] = str(process_time)
    
    # Record metrics
    REQUEST_COUNT.labels(
        method=request.method,
        endpoint=request.url.path,
        status=response.status_code
    ).inc()
    
    REQUEST_LATENCY.observe(process_time)
    
    # Log request
    logger.info(
        "HTTP request processed",
        method=request.method,
        path=request.url.path,
        status_code=response.status_code,
        process_time=process_time,
        client_ip=request.client.host if request.client else None,
    )
    
    return response


@app.exception_handler(Exception)
async def global_exception_handler(request: Request, exc: Exception):
    """Global exception handler"""
    logger.error(
        "Unhandled exception",
        method=request.method,
        path=request.url.path,
        error=str(exc),
        exc_info=True,
    )
    
    return JSONResponse(
        status_code=500,
        content={"detail": "Internal server error"},
    )


@app.exception_handler(HTTPException)
async def http_exception_handler(request: Request, exc: HTTPException):
    """HTTP exception handler"""
    logger.warning(
        "HTTP exception",
        method=request.method,
        path=request.url.path,
        status_code=exc.status_code,
        detail=exc.detail,
    )
    
    return JSONResponse(
        status_code=exc.status_code,
        content={"detail": exc.detail},
    )


# Include API router
app.include_router(api_router, prefix="/api/v1")


@app.get("/")
async def root():
    """Root endpoint"""
    return {
        "message": "SIEM Backend API",
        "version": settings.app_version,
        "status": "running",
    }


@app.get("/health")
async def health_check():
    """Health check endpoint"""
    return {
        "status": "healthy",
        "version": settings.app_version,
        "timestamp": time.time(),
    }


@app.get("/metrics")
async def metrics():
    """Prometheus metrics endpoint"""
    return Response(
        content=generate_latest(),
        media_type=CONTENT_TYPE_LATEST,
    )


if __name__ == "__main__":
    import uvicorn
    
    uvicorn.run(
        "app.main:app",
        host=settings.host,
        port=settings.port,
        reload=settings.debug,
        workers=settings.workers,
    ) 