from fastapi import FastAPI, HTTPException, Depends
from pydantic import BaseModel
from typing import Dict, List, Optional
import os
import uvicorn

from adapters.mlflow_adapter import MLflowAdapter
from adapters.base import ModelRegistration, RegisteredModel

app = FastAPI(title="ML Model Registry API", version="1.0.0")

# Initialize MLflow adapter
mlflow_adapter = MLflowAdapter(
    tracking_uri=os.getenv("MLFLOW_TRACKING_URI", "http://mlflow:5000")
)

class RegisterRequest(BaseModel):
    name: str
    version: str
    framework: str
    artifact_uri: str
    metadata: Dict[str, any] = {}
    metrics: Dict[str, float] = {}
    tags: Dict[str, str] = {}
    lineage_run_id: Optional[str] = None

class PromoteRequest(BaseModel):
    stage: str

@app.get("/health")
async def health():
    return {"status": "healthy"}

@app.post("/api/v1/models/register")
async def register_model(request: RegisterRequest):
    """Register a model in MLflow"""
    try:
        model_reg = ModelRegistration(
            name=request.name,
            version=request.version,
            framework=request.framework,
            artifact_uri=request.artifact_uri,
            metadata=request.metadata,
            metrics=request.metrics,
            tags=request.tags,
            lineage_run_id=request.lineage_run_id
        )
        
        version = mlflow_adapter.register_model(model_reg)
        return {
            "status": "registered",
            "model_id": f"{request.name}/{version}",
            "version": version
        }
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

@app.get("/api/v1/models/{name}/{version}")
async def get_model(name: str, version: str):
    """Get model details"""
    try:
        model = mlflow_adapter.get_model(name, version)
        return model.dict()
    except Exception as e:
        raise HTTPException(status_code=404, detail=f"Model not found: {str(e)}")

@app.get("/api/v1/models")
async def list_models(stage: Optional[str] = None, owner: Optional[str] = None):
    """List all models with optional filters"""
    try:
        filters = {}
        if stage:
            filters["stage"] = stage
        if owner:
            filters["owner"] = owner
        
        models = mlflow_adapter.list_models(filters)
        return {"models": [m.dict() for m in models], "count": len(models)}
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

@app.post("/api/v1/models/{name}/{version}/promote")
async def promote_model(name: str, version: str, request: PromoteRequest):
    """Promote model to a stage"""
    try:
        model_id = f"{name}/{version}"
        mlflow_adapter.promote_model(model_id, request.stage)
        return {"status": "promoted", "stage": request.stage}
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

if __name__ == "__main__":
    uvicorn.run(app, host="0.0.0.0", port=8081)
