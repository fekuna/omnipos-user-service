.PHONY: migrate migrate_down migrate_up migrate_version

# ==============================================================================
# Go migrate postgresql

DB_NAME = omnipos_user_db

migrate_create:
	migrate create -seq -ext=.sql -dir=./migrations ${name}

force:
	migrate -database postgres://omnipos:omnipos@localhost:5433/$(DB_NAME)?sslmode=disable -path migrations force 1

version:
	migrate -database postgres://omnipos:omnipos@localhost:5433/$(DB_NAME)?sslmode=disable -path migrations version

migrate_up:
	migrate -database postgres://omnipos:omnipos@localhost:5433/$(DB_NAME)?sslmode=disable -path migrations up

migrate_down:
	migrate -database postgres://omnipos:omnipos@localhost:5433/$(DB_NAME)?sslmode=disable -path migrations down 1