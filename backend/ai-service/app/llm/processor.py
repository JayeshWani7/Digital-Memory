import json
import logging
from typing import Optional

import openai

from app.models import ProcessedKnowledge, Entity

logger = logging.getLogger(__name__)


class KnowledgeProcessor:
    """Process raw events and extract knowledge using LLMs"""
    
    def __init__(self, config):
        self.config = config
        self.model = config.openai_model
        self.embedding_model = config.openai_embedding_model
        
        # Initialize OpenAI
        openai.api_key = config.openai_api_key
    
    async def process_slack_message(self, event_data: dict, raw_text: str) -> ProcessedKnowledge:
        """Process a Slack message"""
        try:
            summary = await self._generate_summary(raw_text, "slack")
            entities = await self._extract_entities(raw_text, "slack")
            decisions = await self._extract_decisions(raw_text)
            tags = await self._generate_tags(raw_text)
            embedding = await self._generate_embedding(summary)
            
            return ProcessedKnowledge(
                event_id=event_data.get("event_id", ""),
                summary=summary,
                decisions=decisions,
                entities=entities,
                tags=tags,
                raw_text=raw_text,
                embedding=embedding,
                confidence=0.85,
                model_used=self.model
            )
        except Exception as e:
            logger.error(f"Error processing Slack message: {e}")
            raise
    
    async def process_github_event(self, event_data: dict, raw_text: str) -> ProcessedKnowledge:
        """Process a GitHub event (PR, commit, etc.)"""
        try:
            summary = await self._generate_summary(raw_text, "github")
            entities = await self._extract_entities(raw_text, "github")
            decisions = await self._extract_decisions(raw_text)
            tags = await self._generate_tags(raw_text)
            embedding = await self._generate_embedding(summary)
            
            return ProcessedKnowledge(
                event_id=event_data.get("event_id", ""),
                summary=summary,
                decisions=decisions,
                entities=entities,
                tags=tags,
                raw_text=raw_text,
                embedding=embedding,
                confidence=0.90,
                model_used=self.model
            )
        except Exception as e:
            logger.error(f"Error processing GitHub event: {e}")
            raise
    
    async def _generate_summary(self, text: str, source: str) -> str:
        """Generate a concise summary of the text"""
        prompt = f"""
        Summarize the following {source} message in 1-2 sentences. Focus on the key information.
        
        {text}
        
        Summary:
        """
        
        try:
            response = openai.ChatCompletion.create(
                model=self.model,
                messages=[
                    {
                        "role": "system",
                        "content": "You are a technical knowledge extractor. Provide concise, factual summaries."
                    },
                    {"role": "user", "content": prompt}
                ],
                temperature=0.3,
                max_tokens=200
            )
            return response.choices[0].message.content.strip()
        except Exception as e:
            logger.error(f"Error generating summary: {e}")
            return text[:200]  # Fallback to truncated text
    
    async def _extract_entities(self, text: str, source: str) -> list:
        """Extract entities (services, APIs, people, tools) from the text"""
        prompt = f"""
        Extract entities from the following {source} message. Return as JSON array.
        Identify: services, APIs, tools, people, architectural components.
        
        {text}
        
        Return JSON like: [{{"name": "...", "type": "service|api|person|tool|architecture", "context": "..."}}, ...]
        """
        
        try:
            response = openai.ChatCompletion.create(
                model=self.model,
                messages=[
                    {
                        "role": "system",
                        "content": "You are a technical entity extractor. Return valid JSON only."
                    },
                    {"role": "user", "content": prompt}
                ],
                temperature=0.0,
                max_tokens=500
            )
            
            response_text = response.choices[0].message.content.strip()
            
            # Try to parse JSON
            try:
                data = json.loads(response_text)
                return [
                    Entity(
                        name=e.get("name", ""),
                        entity_type=e.get("type", "tool"),
                        context=e.get("context", "")
                    )
                    for e in data if e.get("name")
                ]
            except json.JSONDecodeError:
                logger.warning("Failed to parse entities JSON")
                return []
        except Exception as e:
            logger.error(f"Error extracting entities: {e}")
            return []
    
    async def _extract_decisions(self, text: str) -> list:
        """Extract key decisions from the text"""
        prompt = f"""
        Extract key decisions, action items, or architectural changes from the following text.
        Return as JSON array of strings.
        
        {text}
        
        Return JSON like: ["decision 1", "decision 2", ...]
        """
        
        try:
            response = openai.ChatCompletion.create(
                model=self.model,
                messages=[
                    {
                        "role": "system",
                        "content": "You are a decision extractor. Return valid JSON array of strings."
                    },
                    {"role": "user", "content": prompt}
                ],
                temperature=0.1,
                max_tokens=300
            )
            
            response_text = response.choices[0].message.content.strip()
            
            try:
                return json.loads(response_text)
            except json.JSONDecodeError:
                logger.warning("Failed to parse decisions JSON")
                return []
        except Exception as e:
            logger.error(f"Error extracting decisions: {e}")
            return []
    
    async def _generate_tags(self, text: str) -> list:
        """Generate relevant tags/topics for the text"""
        prompt = f"""
        Generate 3-5 relevant tags for the following technical message.
        Return as JSON array of strings.
        
        {text}
        
        Return JSON like: ["tag1", "tag2", ...]
        """
        
        try:
            response = openai.ChatCompletion.create(
                model=self.model,
                messages=[
                    {
                        "role": "system",
                        "content": "You are a tag generator. Return valid JSON array of relevant technical tags."
                    },
                    {"role": "user", "content": prompt}
                ],
                temperature=0.5,
                max_tokens=100
            )
            
            response_text = response.choices[0].message.content.strip()
            
            try:
                return json.loads(response_text)
            except json.JSONDecodeError:
                logger.warning("Failed to parse tags JSON")
                return []
        except Exception as e:
            logger.error(f"Error generating tags: {e}")
            return []
    
    async def _generate_embedding(self, text: str) -> list:
        """Generate embedding vector for the text"""
        try:
            response = openai.Embedding.create(
                input=text,
                model=self.embedding_model
            )
            return response["data"][0]["embedding"]
        except Exception as e:
            logger.error(f"Error generating embedding: {e}")
            return None
