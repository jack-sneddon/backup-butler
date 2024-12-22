echo "\n******* dry-run ******\n"
go run cmd/main.go -config configs/test_config.json` --dry-run
rm -rf /Users/jack/tmp/2024-11-27-back

echo "\n******* copy *****\n"
go run cmd/main.go -config configs/test_config.json`
du -h /Users/jack/tmp/2024-11-27-back

echo "\n******* copy2 *****\n"
rm -rf /Users/jack/tmp/2024-11-27-back`/Packers
go run cmd/main.go -config configs/test_config.json`

echo "\n******* list versions *****\n"
# List all versions
go run cmd/main.go -config configs/test_config.json` --list-versions

# Show latest version
echo "\n******* show latest versions *****\n"
go run cmd/main.go -config configs/test_config.json` --latest-version
