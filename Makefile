.PHONY: backend frontend test test-backend test-frontend lint seed db-reset db-dump db-restore

backend:
	cd backend && go run ./...

frontend:
	cd frontend && npm install --silent && npm run dev

test: test-backend test-frontend

test-backend:
	cd backend && go test ./...

test-frontend:
	cd frontend && npx vitest run

lint:
	@for s in scripts/check-*.sh; do echo "== $$s"; bash $$s || exit 1; done

seed:
	cd backend && SEED_ON_BOOT=true go run ./... --seed-only

db-reset:
	docker compose down -v db && docker compose up -d db

db-dump:
	@mkdir -p backups && docker compose exec -T db pg_dump -U fieldops fieldops > backups/fieldops-$$(date +%Y%m%d-%H%M%S).sql && ls -1 backups | tail -1

db-restore:
	@test -n "$(FILE)" || (echo "usage: make db-restore FILE=backups/<file>.sql" && exit 1)
	docker compose exec -T db psql -U fieldops -c 'DROP DATABASE IF EXISTS fieldops_restore' -c 'CREATE DATABASE fieldops_restore'
	docker compose exec -T db psql -U fieldops fieldops_restore < $(FILE)
	docker compose exec -T db psql -U fieldops fieldops_restore -c "SELECT relname AS table, n_live_tup AS rows FROM pg_stat_user_tables ORDER BY 1"
