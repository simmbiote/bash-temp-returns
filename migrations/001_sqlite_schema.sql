-- Customer Support Chatbot API - SQLite Schema
-- Migration: 001_sqlite_schema.sql

-- Conversations table
CREATE TABLE conversations (
    id TEXT PRIMARY KEY,
    customer_id TEXT NOT NULL,
    type TEXT NOT NULL CHECK (type IN ('natural_language', 'guided_flow')),
    intent TEXT,
    current_step TEXT,
    flow_id TEXT,
    context TEXT NOT NULL DEFAULT '{}',
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now')),
    completed_at TEXT
);

CREATE INDEX idx_conversations_customer_id ON conversations(customer_id);
CREATE INDEX idx_conversations_created_at ON conversations(created_at);

-- Messages table
CREATE TABLE messages (
    id TEXT PRIMARY KEY,
    conversation_id TEXT NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    role TEXT NOT NULL CHECK (role IN ('user', 'assistant', 'system')),
    content TEXT NOT NULL,
    metadata TEXT NOT NULL DEFAULT '{}',
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX idx_messages_conversation_id ON messages(conversation_id);
CREATE INDEX idx_messages_created_at ON messages(created_at);

-- Return requests table
CREATE TABLE return_requests (
    id TEXT PRIMARY KEY,
    conversation_id TEXT REFERENCES conversations(id) ON DELETE SET NULL,
    customer_id TEXT NOT NULL,
    order_number TEXT NOT NULL,
    items TEXT NOT NULL,
    reason TEXT NOT NULL,
    detailed_reason TEXT,
    photos TEXT,
    refund_method TEXT NOT NULL,
    delivery_method TEXT NOT NULL,
    collection_point TEXT,
    shipping_address TEXT,
    status TEXT NOT NULL DEFAULT 'draft',
    estimated_refund_date TEXT,
    refund_amount_cents INTEGER NOT NULL,
    return_reference TEXT,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX idx_return_requests_customer_id ON return_requests(customer_id);
CREATE INDEX idx_return_requests_order_number ON return_requests(order_number);
CREATE INDEX idx_return_requests_status ON return_requests(status);
CREATE INDEX idx_return_requests_created_at ON return_requests(created_at);

-- Refund requests table (Phase 2)
CREATE TABLE refund_requests (
    id TEXT PRIMARY KEY,
    conversation_id TEXT REFERENCES conversations(id) ON DELETE SET NULL,
    customer_id TEXT NOT NULL,
    order_number TEXT NOT NULL,
    return_reference TEXT,
    refund_amount_cents INTEGER NOT NULL,
    refund_method TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    processed_date TEXT,
    failure_reason TEXT,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX idx_refund_requests_customer_id ON refund_requests(customer_id);
CREATE INDEX idx_refund_requests_order_number ON refund_requests(order_number);
CREATE INDEX idx_refund_requests_status ON refund_requests(status);
CREATE INDEX idx_refund_requests_return_reference ON refund_requests(return_reference);

-- Account actions table (Phase 2)
CREATE TABLE account_actions (
    id TEXT PRIMARY KEY,
    conversation_id TEXT REFERENCES conversations(id) ON DELETE SET NULL,
    customer_id TEXT NOT NULL,
    action_type TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    otp_code TEXT,
    otp_expires_at TEXT,
    otp_attempts INTEGER DEFAULT 0,
    action_data TEXT NOT NULL DEFAULT '{}',
    verified_at TEXT,
    completed_at TEXT,
    failure_reason TEXT,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX idx_account_actions_customer_id ON account_actions(customer_id);
CREATE INDEX idx_account_actions_status ON account_actions(status);
CREATE INDEX idx_account_actions_action_type ON account_actions(action_type);
CREATE INDEX idx_account_actions_created_at ON account_actions(created_at);

-- Flows table
CREATE TABLE flows (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    intent TEXT NOT NULL,
    version INTEGER NOT NULL DEFAULT 1,
    steps TEXT NOT NULL,
    is_active INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX idx_flows_intent ON flows(intent);
CREATE INDEX idx_flows_is_active ON flows(is_active);
CREATE UNIQUE INDEX idx_flows_active_intent ON flows(intent) WHERE is_active = 1;

-- Audit log for tracking changes
CREATE TABLE audit_logs (
    id TEXT PRIMARY KEY,
    entity_type TEXT NOT NULL,
    entity_id TEXT NOT NULL,
    action TEXT NOT NULL,
    user_id TEXT,
    changes TEXT,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX idx_audit_logs_entity ON audit_logs(entity_type, entity_id);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at);
