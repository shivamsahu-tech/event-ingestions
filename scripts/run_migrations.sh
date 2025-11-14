set -e  

CONTAINER_NAME="analytics_db"
DB_USER="postgres"
DB_NAME="analytics"

echo "Running database migrations..."

for file in migrations/*.sql; do
  echo "Applying migration: $file"
  docker exec -i $CONTAINER_NAME psql -U $DB_USER -d $DB_NAME < $file
done

echo "All migrations applied successfully!"
