import asyncio
import json
import logging
import time
from datetime import datetime
from typing import Optional, Dict, Any

import redis.asyncio as aioredis
import psycopg2
from psycopg2.extras import RealDictCursor

logger = logging.getLogger(__name__)

# Global database connection pool
_db_connection = None
_redis_connection = None


async def get_db_connection(database_url: str):
    """Get or create database connection"""
    global _db_connection
    
    if _db_connection is None:
        try:
            _db_connection = psycopg2.connect(database_url)
            _db_connection.autocommit = True
            logger.info("Database connection established")
        except Exception as e:
            logger.error(f"Failed to connect to database: {e}")
            raise
    
    return _db_connection


async def close_db_connection():
    """Close database connection"""
    global _db_connection
    
    if _db_connection:
        _db_connection.close()
        _db_connection = None
        logger.info("Database connection closed")


async def get_redis_connection(redis_url: str):
    """Get or create Redis connection"""
    global _redis_connection
    
    if _redis_connection is None:
        try:
            _redis_connection = aioredis.from_url(redis_url)
            await _redis_connection.ping()
            logger.info("Redis connection established")
        except Exception as e:
            logger.error(f"Failed to connect to Redis: {e}")
            raise
    
    return _redis_connection


async def close_redis_connection():
    """Close Redis connection"""
    global _redis_connection
    
    if _redis_connection:
        await _redis_connection.close()
        _redis_connection = None
        logger.info("Redis connection closed")


async def store_knowledge(
    event_id: str,
    summary: str,
    raw_text: str,
    entities: list,
    decisions: list,
    tags: list,
    embedding: Optional[list],
    confidence: float,
    model: str
) -> str:
    """Store processed knowledge in database"""
    conn = await get_db_connection(_db_connection)
    
    with conn.cursor(cursor_factory=RealDictCursor) as cur:
        # Insert knowledge
        query = """
            INSERT INTO knowledge 
            (event_id, summary, raw_text, decisions, tags, confidence, model_used)
            VALUES (%s, %s, %s, %s, %s, %s, %s)
            RETURNING id
        """
        
        cur.execute(
            query,
            (
                event_id,
                summary,
                raw_text,
                json.dumps(decisions),
                json.dumps(tags),
                confidence,
                model
            )
        )
        
        knowledge_id = cur.fetchone()["id"]
        
        # Store entities relationship if embedding is ready
        if embedding:
            embedding_query = """
                UPDATE knowledge
                SET embedding = %s, updated_at = NOW()
                WHERE id = %s
            """
            cur.execute(embedding_query, (embedding, knowledge_id))
        
        return knowledge_id


async def get_event_by_id(event_id: str) -> Optional[Dict[str, Any]]:
    """Retrieve event details from database"""
    conn = await get_db_connection(_db_connection)
    
    with conn.cursor(cursor_factory=RealDictCursor) as cur:
        query = """
            SELECT id, source, source_id, event_type, raw_data, author, channel, received_at
            FROM events
            WHERE id = %s
        """
        cur.execute(query, (event_id,))
        result = cur.fetchone()
        
        if result:
            result["raw_data"] = json.loads(result["raw_data"])
        
        return result


async def update_event_status(event_id: str, status: str):
    """Update event processing status"""
    conn = await get_db_connection(_db_connection)
    
    with conn.cursor() as cur:
        query = """
            UPDATE events
            SET processing_status = %s, processed_at = NOW()
            WHERE id = %s
        """
        cur.execute(query, (status, event_id))
        conn.commit()


async def record_processing_error(
    event_id: str,
    service: str,
    error_type: str,
    error_message: str,
    stack_trace: str = ""
):
    """Record a processing error"""
    conn = await get_db_connection(_db_connection)
    
    with conn.cursor() as cur:
        query = """
            INSERT INTO processing_errors
            (event_id, service, error_type, error_message, stack_trace, occurred_at)
            VALUES (%s, %s, %s, %s, %s, NOW())
        """
        cur.execute(query, (event_id, service, error_type, error_message, stack_trace))
        conn.commit()
