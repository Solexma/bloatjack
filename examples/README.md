# Bloatjack Examples

This directory contains example configurations and test cases for demonstrating Bloatjack's capabilities.

## Microservices Stack Example

The `microservices-stack.yml` file provides a Docker Compose configuration with common service types that Bloatjack can optimize:

- **Database** (PostgreSQL): Demonstrates memory cap rules
- **API server** (Node.js): Shows optimization for web services
- **Cache** (Redis): Tests memory pattern analysis

### Usage

1. **Start the example stack**:

   ```bash
   docker-compose -f examples/microservices-stack.yml up -d
   ```

2. **Generate some load** (optional):

   Make the load generation scripts executable:

   ```bash
   chmod +x examples/generate-postgres-load.sh
   chmod +x examples/generate-redis-load.sh
   ```

   Run the scripts to populate the database and cache:

   ```bash
   ./examples/generate-postgres-load.sh
   ./examples/generate-redis-load.sh
   ```

   *Note: The example API service (`bloatjack-example-api`) is a placeholder and does not have specific load generation steps.*

3. **Run Bloatjack to analyze**:

   ```bash
   bloatjack scan
   ```

4. **Apply optimizations**:

   ```bash
   bloatjack tune
   ```

5. **Verify changes**:

   ```bash
   docker-compose -f examples/microservices-stack.yml config
   ```

## Testing Rules

The example stack works with the built-in rules from `internal/rules/*.yml`. The services have labels that match the rule selectors:

```yaml
# Example rule from internal/rules/memory.yml
- id: mem-cap-db@1.0.0
  priority: 80
  match: {kind: db}
  if: peak_mem_mb > 800
  set:
    mem_limit: '{peak_mem_mb*1.25}m'
```

## Adding New Examples

Feel free to contribute additional example stacks by:

1. Creating a new YAML file in this directory
2. Adding appropriate labels to services for Bloatjack to recognize
3. Updating this README with usage instructions

## Running Unit Tests

To run unit tests for the rules engine:

```bash
go test github.com/Solexma/bloatjack/internal/rules
```
