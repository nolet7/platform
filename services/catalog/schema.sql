-- Platform Catalog Database Schema

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Main entities table
CREATE TABLE IF NOT EXISTS entities (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    type VARCHAR(50) NOT NULL CHECK (type IN ('service', 'model', 'dataset', 'pipeline', 'environment')),
    name VARCHAR(255) NOT NULL,
    owner_team VARCHAR(100) NOT NULL,
    owner_email VARCHAR(255) NOT NULL,
    tier VARCHAR(20) CHECK (tier IN ('critical', 'high', 'medium', 'low')),
    data_classification VARCHAR(50) CHECK (data_classification IN ('public', 'internal', 'confidential', 'restricted')),
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    created_by VARCHAR(255),
    updated_by VARCHAR(255),
    CONSTRAINT unique_name_type UNIQUE (name, type)
);

-- Entity relationships
CREATE TABLE IF NOT EXISTS entity_relationships (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    source_entity_id UUID REFERENCES entities(id) ON DELETE CASCADE,
    target_entity_id UUID REFERENCES entities(id) ON DELETE CASCADE,
    relationship_type VARCHAR(50) NOT NULL,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT NOW(),
    CONSTRAINT unique_relationship UNIQUE (source_entity_id, target_entity_id, relationship_type)
);

-- Scorecards for compliance/readiness tracking
CREATE TABLE IF NOT EXISTS scorecards (
    entity_id UUID REFERENCES entities(id) ON DELETE CASCADE,
    category VARCHAR(50) NOT NULL,
    score INTEGER CHECK (score >= 0 AND score <= 100),
    checks JSONB DEFAULT '[]',
    last_evaluated TIMESTAMP DEFAULT NOW(),
    PRIMARY KEY (entity_id, category)
);

-- Audit log
CREATE TABLE IF NOT EXISTS audit_events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    event_type VARCHAR(100) NOT NULL,
    entity_id UUID,
    actor VARCHAR(255) NOT NULL,
    action VARCHAR(100) NOT NULL,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX idx_entities_type ON entities(type);
CREATE INDEX idx_entities_owner ON entities(owner_team);
CREATE INDEX idx_entities_tier ON entities(tier);
CREATE INDEX idx_entities_metadata ON entities USING GIN(metadata);
CREATE INDEX idx_entity_relationships_source ON entity_relationships(source_entity_id);
CREATE INDEX idx_entity_relationships_target ON entity_relationships(target_entity_id);
CREATE INDEX idx_audit_events_entity ON audit_events(entity_id);
CREATE INDEX idx_audit_events_created ON audit_events(created_at);

-- Updated timestamp trigger
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_entities_updated_at BEFORE UPDATE ON entities
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Sample data for testing
INSERT INTO entities (type, name, owner_team, owner_email, tier, data_classification, metadata)
VALUES 
    ('service', 'payment-api', 'payments-team', 'payments@company.com', 'critical', 'confidential', 
     '{"language": "go", "runtime": "kubernetes"}'::jsonb),
    ('service', 'user-service', 'identity-team', 'identity@company.com', 'high', 'internal',
     '{"language": "python", "runtime": "kubernetes"}'::jsonb),
    ('model', 'fraud-detector', 'ml-team', 'ml@company.com', 'high', 'confidential',
     '{"framework": "sklearn", "version": "1.0.0"}'::jsonb)
ON CONFLICT (name, type) DO NOTHING;
