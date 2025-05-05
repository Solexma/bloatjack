#!/bin/bash
# Generates load on the bloatjack-example-redis container

echo "Generating load on Redis..."

# Simulate writing a moderate number of keys
echo "Writing keys..."
for i in {1..5000}; do
  # Use pipeline for potentially better performance if needed, but individual commands are fine for example
  docker exec bloatjack-example-redis redis-cli SET "key-$i" "value-$RANDOM-$RANDOM-$i" > /dev/null
  if (( i % 500 == 0 )); then
    echo -n "."
  fi
done
echo ""

# Simulate reading a subset of keys randomly
echo "Reading keys..."
for i in {1..2000}; do
  docker exec bloatjack-example-redis redis-cli GET "key-$((RANDOM % 5000 + 1))" > /dev/null
   if (( i % 200 == 0 )); then
    echo -n "."
  fi
done
echo ""

echo "Redis load generation complete." 