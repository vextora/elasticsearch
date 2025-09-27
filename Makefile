DB_SERVICE=ecs_db
ES_SERVICE=elasticsearch
APP_NAME=ecs_app

.PHONY: wait-db wait-es

wait-db:
	@echo "Waiting for Postgres..."
	@until docker exec $(DB_SERVICE) pg_isready -U $$POSTGRES_USER -d $$POSTGRES_DB; do \
		sleep 2; \
	done
	@echo "Postgres ready."

wait-es:
	@echo "Waiting for Elasticsearch..."
	@until curl -s http://localhost:9200/_cluster/health | grep -q '"status":"green"'; do \
		sleep 2; \
	done
	@echo "Elasticsearch ready."

.PHONY: dev
dev:
	docker-compose up --build

.PHONY: seed
seed: wait-db
	@echo "Seeding database..."
	# Cek apakah binary sudah ada di container, kalau tidak ada build dulu
	@if ! docker exec $(APP_NAME) test -f tmp/main; then \
		echo "Building binary inside container..."; \
		docker exec $(APP_NAME) go build -o tmp/main cmd/app/main.go; \
	fi
	docker exec -it $(APP_NAME) ./tmp/main seed

.PHONY: sync
sync: wait-db wait-es
	@echo "Syncing to Elasticsearch..."
	docker exec -it $(APP_NAME) ./tmp/main sync
