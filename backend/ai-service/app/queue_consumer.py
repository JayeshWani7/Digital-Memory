import asyncio
import json
import logging
from typing import Dict, Any

import redis.asyncio as aioredis

from app.database import (
    get_event_by_id,
    store_knowledge,
    update_event_status,
    record_processing_error
)
from app.models import QueueEvent

logger = logging.getLogger(__name__)


class QueueConsumer:
    """Consume events from Redis queue and process them"""
    
    def __init__(self, config, knowledge_processor):
        self.config = config
        self.knowledge_processor = knowledge_processor
        self.redis = None
        self.running = False
        self.metrics = {
            "processed": 0,
            "failed": 0,
            "pending": 0
        }
    
    async def start(self):
        """Start consuming events from queue"""
        try:
            self.redis = aioredis.from_url(self.config.redis_url)
            await self.redis.ping()
            logger.info("Queue consumer connected to Redis")
            self.running = True
            
            # Start consuming from multiple streams
            await self._consume_events()
        except Exception as e:
            logger.error(f"Failed to start queue consumer: {e}")
            self.running = False
            raise
    
    async def _consume_events(self):
        """Consume events from Redis streams"""
        streams = [
            f"events.slack.message",
            f"events.github.pr_created",
            f"events.github.pr_updated",
            f"events.github.commit"
        ]
        
        last_ids = {stream: '0' for stream in streams}
        
        while self.running:
            try:
                # Read from multiple streams
                results = await self.redis.xread(last_ids, block=1000)
                
                if not results:
                    continue
                
                for stream, messages in results:
                    for msg_id, data in messages:
                        try:
                            event_json = data[b'event'].decode('utf-8')
                            event_data = json.loads(event_json)
                            
                            await self._process_event(event_data)
                            self.metrics["processed"] += 1
                            
                            last_ids[stream.decode()] = msg_id
                        except Exception as e:
                            logger.error(f"Error processing message: {e}")
                            self.metrics["failed"] += 1
                
            except asyncio.CancelledError:
                logger.info("Queue consumer cancelled")
                break
            except Exception as e:
                logger.error(f"Error in event consumer loop: {e}")
                await asyncio.sleep(5)  # Backoff on error
    
    async def _process_event(self, event_data: Dict[str, Any]):
        """Process a single event"""
        event_id = event_data.get("event_id")
        source = event_data.get("source")
        event_type = event_data.get("event_type")
        
        try:
            # Update status to processing
            await update_event_status(event_id, "processing")
            
            # Fetch full event details
            event_details = await get_event_by_id(event_id)
            if not event_details:
                logger.warning(f"Event {event_id} not found")
                return
            
            raw_text = self._extract_raw_text(event_details)
            
            # Process based on source
            if source == "slack":
                knowledge = await self.knowledge_processor.process_slack_message(
                    event_data,
                    raw_text
                )
            elif source == "github":
                knowledge = await self.knowledge_processor.process_github_event(
                    event_data,
                    raw_text
                )
            else:
                logger.warning(f"Unknown source: {source}")
                return
            
            # Store knowledge
            knowledge_id = await store_knowledge(
                event_id=event_id,
                summary=knowledge.summary,
                raw_text=knowledge.raw_text,
                entities=[e.to_dict() for e in knowledge.entities],
                decisions=knowledge.decisions,
                tags=knowledge.tags,
                embedding=knowledge.embedding,
                confidence=knowledge.confidence,
                model=knowledge.model_used
            )
            
            logger.info(f"Knowledge stored for event {event_id}: {knowledge_id}")
            
            # Update event status
            await update_event_status(event_id, "completed")
            
        except Exception as e:
            logger.error(f"Error processing event {event_id}: {e}")
            await update_event_status(event_id, "failed")
            await record_processing_error(
                event_id=event_id,
                service="ai-service",
                error_type="processing_error",
                error_message=str(e)
            )
    
    def _extract_raw_text(self, event: Dict[str, Any]) -> str:
        """Extract raw text from event based on source"""
        raw_data = event.get("raw_data", {})
        source = event.get("source")
        
        if source == "slack":
            return raw_data.get("text", "")
        elif source == "github":
            # GitHub: combine title and body
            title = raw_data.get("pull_request", {}).get("title", "")
            body = raw_data.get("pull_request", {}).get("body", "")
            message = raw_data.get("head_commit", {}).get("message", "")
            
            return f"{title} {body} {message}".strip()
        
        return ""
    
    async def close(self):
        """Close the queue consumer"""
        self.running = False
        if self.redis:
            await self.redis.close()
        logger.info("Queue consumer closed")
    
    def is_connected(self) -> bool:
        """Check if consumer is connected"""
        return self.redis is not None
    
    def get_metrics(self) -> Dict[str, int]:
        """Get consumer metrics"""
        return self.metrics.copy()
