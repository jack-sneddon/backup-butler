#!/bin/bash
# test_logging.sh

# Create test configs with different log levels
cat > configs/test-quiet.yaml <<EOF
source_directory: "/Users/jack/tmp/2024-11-27"
target_directory: "/Users/jack/tmp/2024-11-27-backup"
folders_to_backup:
  - "Packers"
  - "Photos-001"
  - "photos"
log_level: "quiet"
EOF

cat > configs/test-verbose.yaml <<EOF
source_directory: "/Users/jack/tmp/2024-11-27"
target_directory: "/Users/jack/tmp/2024-11-27-backup"
folders_to_backup:
  - "Packers"
  - "Photos-001"
  - "photos"
log_level: "verbose"
EOF

cat > configs/test-debug.yaml <<EOF
source_directory: "/Users/jack/tmp/2024-11-27"
target_directory: "/Users/jack/tmp/2024-11-27-backup"
folders_to_backup:
  - "Packers"
  - "Photos-001"
  - "photos"
log_level: "debug"
EOF

echo "Testing with quiet logging..."
./backup-butler -config configs/test-quiet.yaml

echo -e "\nTesting with verbose logging..."
./backup-butler -config configs/test-verbose.yaml

echo -e "\nTesting with debug logging..."
./backup-butler -config configs/test-debug.yaml
