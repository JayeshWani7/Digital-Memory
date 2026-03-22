from dataclasses import dataclass, field
from typing import List, Dict, Any, Optional
from datetime import datetime


@dataclass
class Entity:
    """Extracted entity"""
    name: str
    entity_type: str  # service, api, person, tool, decision, architecture
    context: str = ""
    confidence: float = 0.0
    
    def to_dict(self) -> Dict[str, Any]:
        return {
            "name": self.name,
            "type": self.entity_type,
            "context": self.context,
            "confidence": self.confidence
        }


@dataclass
class ProcessedKnowledge:
    """Result of LLM processing"""
    event_id: str
    summary: str
    decisions: List[str] = field(default_factory=list)
    entities: List[Entity] = field(default_factory=list)
    tags: List[str] = field(default_factory=list)
    confidence: float = 0.0
    raw_text: str = ""
    model_used: str = "gpt-3.5-turbo"
    embedding: Optional[List[float]] = None
    
    def to_dict(self) -> Dict[str, Any]:
        return {
            "event_id": self.event_id,
            "summary": self.summary,
            "decisions": self.decisions,
            "entities": [e.to_dict() for e in self.entities],
            "tags": self.tags,
            "confidence": self.confidence,
            "raw_text": self.raw_text,
            "model_used": self.model_used,
            "embedding_dimension": len(self.embedding) if self.embedding else 0
        }


@dataclass
class QueueEvent:
    """Event from queue"""
    event_id: str
    source: str  # slack, github
    event_type: str  # message, pr_created, commit, etc.
    timestamp: datetime
    data: Dict[str, Any]
