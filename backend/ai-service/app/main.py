import asyncio
import json
import logging
import os
from contextlib import asynccontextmanager

from dotenv import load_dotenv
from fastapi import FastAPI, HTTPException
from fastapi.responses import JSONResponse

from app.config import Config
from app.database import get_db_connection, close_db_connection, init_database
from app.queue_consumer import QueueConsumer
from app.llm.processor import KnowledgeProcessor

# Load environment variables — try multiple paths so it works from any CWD
import pathlib

def _load_env() -> None:
    """Search for .env in several locations and load the first one found."""
    # Absolute path of the directory containing this file (app/)
    here = pathlib.Path(__file__).resolve().parent
    candidates = [
        here / ".env",            # app/.env
        here.parent / ".env",     # ai-service/.env  ← recommended placement
        here.parent.parent / ".env",   # backend/.env
        here.parent.parent.parent / ".env",  # Digital-Memory/.env (project root)
    ]
    for path in candidates:
        if path.exists():
            load_dotenv(str(path))
            print(f"[startup] Loaded .env from: {path}")
            return
    print("[startup] WARNING: No .env file found. Relying on system environment variables.")

# Must be called BEFORE Config() or any os.getenv() reads
_load_env()

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

# Global state
config = None
queue_consumer = None
knowledge_processor = None
consumer_task = None


@asynccontextmanager
async def lifespan(app: FastAPI):
    """Startup and shutdown logic"""
    global config, queue_consumer, knowledge_processor, consumer_task
    
    logger.info("Starting AI Service...")
    
    # Initialize configuration
    config = Config()

    # Register the DB URL in the database module so all helpers can use it
    init_database(config.database_url)

    # Initialize database connection
    try:
        await get_db_connection(config.database_url)
        logger.info("Database connection established")
    except Exception as e:
        logger.error(f"Failed to connect to database: {e}")
        raise
    
    # Initialize processors
    try:
        knowledge_processor = KnowledgeProcessor(config)
        logger.info("Knowledge processor initialized")
    except Exception as e:
        logger.error(f"Failed to initialize knowledge processor: {e}")
        raise
    
    # Initialize queue consumer
    try:
        queue_consumer = QueueConsumer(config, knowledge_processor)
        logger.info("Queue consumer initialized")
    except Exception as e:
        logger.error(f"Failed to initialize queue consumer: {e}")
        raise
    
    # Start consumer in background
    consumer_task = asyncio.create_task(queue_consumer.start())
    logger.info("Queue consumer started")
    
    yield  # Application runs
    
    # Shutdown
    logger.info("Shutting down AI Service...")
    if consumer_task:
        consumer_task.cancel()
        try:
            await consumer_task
        except asyncio.CancelledError:
            pass
    
    if queue_consumer:
        await queue_consumer.close()
    
    await close_db_connection()
    logger.info("AI Service shut down complete")


# Create FastAPI app
app = FastAPI(
    title="Digital Memory - AI Processing Service",
    description="LLM-based knowledge extraction and embedding generation",
    version="1.0.0",
    lifespan=lifespan
)


@app.get("/health")
async def health_check():
    """Health check endpoint"""
    return {
        "status": "healthy",
        "service": "ai-service",
        "timestamp": json.dumps({"_": 0}, default=str)  # UTC timestamp
    }


@app.get("/status")
async def status():
    """Get service status"""
    if not queue_consumer or not knowledge_processor:
        return JSONResponse(
            {"status": "initializing"},
            status_code=503
        )
    
    return {
        "service": "ai-service",
        "status": "ready",
        "processor": {
            "model": knowledge_processor.model,
            "embedding_model": knowledge_processor.embedding_model
        },
        "queue": {
            "connected": queue_consumer.is_connected()
        }
    }


@app.get("/metrics")
async def metrics():
    """Get service metrics"""
    if not queue_consumer:
        return {"processed": 0, "failed": 0, "pending": 0}
    
    return queue_consumer.get_metrics()


@app.exception_handler(Exception)
async def exception_handler(request, exc):
    """Global exception handler"""
    logger.error(f"Unhandled exception: {exc}", exc_info=True)
    return JSONResponse(
        {"error": "Internal server error"},
        status_code=500
    )


if __name__ == "__main__":
    import uvicorn
    
    port = int(os.getenv("PORT", "8002"))
    uvicorn.run(
        app,
        host="0.0.0.0",
        port=port,
        log_level="info"
    )
