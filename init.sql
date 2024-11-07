-- Create the debug_logs table for logging errors
CREATE TABLE IF NOT EXISTS debug_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMP DEFAULT NOW(),
    log_level VARCHAR(50),
    error JSONB,
    message TEXT,
    context JSONB
);

-- Example insert to test table
INSERT INTO debug_logs (log_level, error, message, context)
VALUES 
    ('ERROR', '{"code": 500, "message": "Server exited", "details": "A panic occurred and the server has exited.", "origin": "api.main.handler"}', 
    'An example error message for testing.', 
    '{"param": "value"}');
