.PHONY: build dev release clean

# Build frontend then compile Go binary
build:
	cd frontend && npm ci && npm run build
	go build -o photoorg .

# Run backend only (frontend dev server handles UI at :3011)
dev:
	go run .

# Dry-run GoReleaser (validates config, no publish)
release-dry:
	goreleaser release --snapshot --clean

# Tag and push to trigger GitHub Actions release
release:
	@if [ -z "$(VERSION)" ]; then echo "Usage: make release VERSION=x.y.z"; exit 1; fi
	git tag -a v$(VERSION) -m "Release v$(VERSION)"
	git push origin v$(VERSION)

clean:
	rm -f photoorg photoorg.exe
	rm -rf frontend/dist dist
