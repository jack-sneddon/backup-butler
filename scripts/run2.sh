#!/bin/bash

# First clean any previous backups to see full comparison process
rm -rf /Users/jack/tmp/2024-11-27-backup

# Run backup with debug logging
./backup-butler -config configs/config.yaml

# Run again to see comparison strategies in action
./backup-butler -config configs/config.yaml
