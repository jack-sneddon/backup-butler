#!/bin/zsh

# Define the project root and output zip file
PROJECT_ROOT=$(pwd)
OUTPUT_ZIP="backup-butler-source.zip"

# Create the zip archive, excluding unwanted files and directories
zip -r $OUTPUT_ZIP $PROJECT_ROOT \
    -x "*.out" \
    -x "*.exe" \
    -x "*.dll" \
    -x "*.so" \
    -x "*.a" \
    -x "*.dylib" \
    -x "*.pyc" \
    -x "*.o" \
    -x "*.swp" \
    -x "utilities/*" \
    -x "*/out/*" \
    -x "*.log" \
    -x "*.tmp" \
    -x "vendor/*" \
    -x "node_modules/*"

# Print a success message
echo "Project zipped successfully as $OUTPUT_ZIP"