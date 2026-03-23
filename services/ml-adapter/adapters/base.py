from abc import ABC, abstractmethod
from dataclasses import dataclass
from typing import Dict, List, Optional
from datetime import datetime

@dataclass
class ModelRegistration:
    name: str
    version: str
    framework: str
    artifact_uri: str
    metadata: Dict[str, any]
    metrics: Dict[str, float]
    tags: Dict[str, str]
    lineage_run_id: Optional[str] = None

@dataclass
class RegisteredModel:
    id: str
    name: str
    version: str
    stage: str
    artifact_uri: str
    metrics: Dict[str, float]
    metadata: Dict[str, any]
    created_at: datetime
    updated_by: str
    
    def dict(self):
        return {
            "id": self.id,
            "name": self.name,
            "version": self.version,
            "stage": self.stage,
            "artifact_uri": self.artifact_uri,
            "metrics": self.metrics,
            "metadata": self.metadata,
            "created_at": self.created_at.isoformat() if self.created_at else None,
            "updated_by": self.updated_by
        }

class ModelRegistryAdapter(ABC):
    """Abstract base for model registry adapters"""
    
    @abstractmethod
    def register_model(self, model: ModelRegistration) -> str:
        pass
    
    @abstractmethod
    def get_model(self, name: str, version: str) -> RegisteredModel:
        pass
    
    @abstractmethod
    def promote_model(self, model_id: str, stage: str) -> None:
        pass
    
    @abstractmethod
    def list_models(self, filters: Optional[Dict] = None) -> List[RegisteredModel]:
        pass
