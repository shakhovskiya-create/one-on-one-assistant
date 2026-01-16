-- BPMN process instance tracking
CREATE TABLE IF NOT EXISTS bpmn_instances (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    camunda_id VARCHAR(100) UNIQUE NOT NULL,
    definition_id VARCHAR(200),
    process_key VARCHAR(100) NOT NULL,
    business_key VARCHAR(200),
    status VARCHAR(50) DEFAULT 'active',
    variables JSONB DEFAULT '{}'::jsonb,
    started_by UUID REFERENCES employees(id),
    started_at TIMESTAMPTZ DEFAULT NOW(),
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- BPMN task history
CREATE TABLE IF NOT EXISTS bpmn_task_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    camunda_task_id VARCHAR(100) NOT NULL,
    instance_id UUID REFERENCES bpmn_instances(id),
    task_name VARCHAR(255),
    task_key VARCHAR(100),
    assignee_id UUID REFERENCES employees(id),
    status VARCHAR(50) DEFAULT 'pending',
    variables JSONB DEFAULT '{}'::jsonb,
    claimed_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_bpmn_instances_status ON bpmn_instances(status);
CREATE INDEX IF NOT EXISTS idx_bpmn_instances_process_key ON bpmn_instances(process_key);
CREATE INDEX IF NOT EXISTS idx_bpmn_instances_business_key ON bpmn_instances(business_key);
CREATE INDEX IF NOT EXISTS idx_bpmn_task_history_instance ON bpmn_task_history(instance_id);
CREATE INDEX IF NOT EXISTS idx_bpmn_task_history_assignee ON bpmn_task_history(assignee_id);
