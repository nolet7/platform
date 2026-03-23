from typing import Dict, List, Optional
from datetime import datetime
import mlflow
from mlflow.tracking import MlflowClient

from .base import ModelRegistryAdapter, ModelRegistration, RegisteredModel

class MLflowAdapter(ModelRegistryAdapter):
    """MLflow Registry Adapter"""
    
    def __init__(self, tracking_uri: str):
        self.client = MlflowClient(tracking_uri)
        mlflow.set_tracking_uri(tracking_uri)
    
    def register_model(self, model: ModelRegistration) -> str:
        # Create registered model if doesn't exist
        try:
            self.client.create_registered_model(
                name=model.name,
                tags=model.tags,
                description=model.metadata.get("description", "")
            )
        except mlflow.exceptions.RestException:
            pass
        
        # Create model version
        version = self.client.create_model_version(
            name=model.name,
            source=model.artifact_uri,
            run_id=model.lineage_run_id,
            tags=model.tags
        )
        
        # Store metrics as tags
        for key, value in model.metrics.items():
            self.client.set_model_version_tag(
                name=model.name,
                version=version.version,
                key=f"metric.{key}",
                value=str(value)
            )
        
        # Store metadata
        for key, value in model.metadata.items():
            self.client.set_model_version_tag(
                name=model.name,
                version=version.version,
                key=f"metadata.{key}",
                value=str(value)
            )
        
        return version.version
    
    def get_model(self, name: str, version: str) -> RegisteredModel:
        mv = self.client.get_model_version(name=name, version=version)
        
        metrics = {}
        metadata = {}
        for key, value in mv.tags.items():
            if key.startswith("metric."):
                metrics[key.replace("metric.", "")] = float(value)
            elif key.startswith("metadata."):
                metadata[key.replace("metadata.", "")] = value
        
        return RegisteredModel(
            id=f"{name}/{version}",
            name=name,
            version=version,
            stage=mv.current_stage,
            artifact_uri=mv.source,
            metrics=metrics,
            metadata=metadata,
            created_at=datetime.fromtimestamp(mv.creation_timestamp / 1000),
            updated_by=mv.user_id
        )
    
    def promote_model(self, model_id: str, stage: str) -> None:
        name, version = model_id.split("/")
        self.client.transition_model_version_stage(
            name=name,
            version=version,
            stage=stage,
            archive_existing_versions=True
        )
    
    def list_models(self, filters: Optional[Dict] = None) -> List[RegisteredModel]:
        models = []
        for rm in self.client.search_registered_models():
            for mv in self.client.search_model_versions(f"name='{rm.name}'"):
                model = self.get_model(rm.name, mv.version)
                
                # Apply filters
                if filters:
                    if "stage" in filters and model.stage != filters["stage"]:
                        continue
                    if "owner" in filters and model.metadata.get("owner") != filters["owner"]:
                        continue
                
                models.append(model)
        
        return models
