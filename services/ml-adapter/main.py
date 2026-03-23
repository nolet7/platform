from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
from typing import Dict
import os

app = FastAPI(title="ML Model Registry API")

# Simple in-memory storage
models = {}

class ModelRequest(BaseModel):
    name: str
    version: str
    framework: str
    artifact_uri: str
    metadata: Dict[str, str] = {}
    metrics: Dict[str, float] = {}
    tags: Dict[str, str] = {}

@app.get("/health")
def health():
    return {"status": "healthy"}

@app.post("/api/v1/models/register")
def register_model(model: ModelRequest):
    model_id = f"{model.name}/{model.version}"
    models[model_id] = model.dict()
    return {"status": "registered", "model_id": model_id}

@app.get("/api/v1/models")
def list_models():
    return {"models": list(models.values()), "count": len(models)}

@app.get("/api/v1/models/{name}/{version}")
def get_model(name: str, version: str):
    model_id = f"{name}/{version}"
    if model_id not in models:
        raise HTTPException(404, "Model not found")
    return models[model_id]
