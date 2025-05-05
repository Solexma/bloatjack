#!/bin/bash
# Generates load on the bloatjack-example-postgres container

PG_USER="bloatjack"
PG_DB="bloatjack_demo"
CONTAINER_NAME="bloatjack-example-postgres"

echo "Generating load on PostgreSQL ($CONTAINER_NAME)..."

# Use heredoc to pass SQL commands to psql inside the container
docker exec -i $CONTAINER_NAME psql -U $PG_USER -d $PG_DB <<EOF
-- Create a table if it doesn't exist
CREATE TABLE IF NOT EXISTS load_test (
    id SERIAL PRIMARY KEY,
    data TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Insert a significant amount of data to increase memory usage
-- Using md5(random()::text) generates somewhat realistic text data
INSERT INTO load_test (data)
SELECT md5(random()::text) FROM generate_series(1, 20000);

-- Perform some reads and basic analysis
SELECT count(*) FROM load_test;
SELECT * FROM load_test ORDER BY random() LIMIT 10;

-- Simulate some updates
UPDATE load_test SET data = data || ' updated' WHERE id % 100 = 0;

-- Run analyze to update statistics, which can consume resources
ANALYZE load_test;

-- Optionally, run vacuum for more intensive I/O and memory
-- VACUUM FULL load_test; -- Uncomment if heavier load is needed

-- Clean up older data (optional)
-- DELETE FROM load_test WHERE created_at < NOW() - INTERVAL '1 hour';
EOF

# Check exit status of docker exec
if [ $? -eq 0 ]; then
  echo "PostgreSQL load generation successful."
else
  echo "Error generating PostgreSQL load. Check container logs."
  exit 1
fi 