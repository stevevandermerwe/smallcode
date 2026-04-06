#!/bin/bash

# Configuration
VM_NAME="my-debian-vm"
REMOTE_USER="admin"
REMOTE_PASS="admin"
SOURCE_FILE="./dist/smallcode-linux-arm64"
TARGET_BIN_DIR="\$HOME/bin"
TARGET_FILENAME="smallcode"

# 1. Validate local file exists
if [ ! -f "$SOURCE_FILE" ]; then
    echo "Error: $SOURCE_FILE not found."
    exit 1
fi

# 2. Get VM IP
VM_IP=$(tart ip "$VM_NAME")
if [ -z "$VM_IP" ]; then
    echo "Error: VM '$VM_NAME' is not running or IP not found."
    exit 1
fi

echo "Deploying to $VM_IP..."

# 3. The SSH Pipe
# - Creates the bin directory
# - Transfers and renames the file
# - Sets executable bit
# - Safely updates .bashrc PATH
cat "$SOURCE_FILE" | sshpass -p "$REMOTE_PASS" ssh -o StrictHostKeyChecking=no "$REMOTE_USER@$VM_IP" "
    mkdir -p $TARGET_BIN_DIR
    
    # Save the piped stream to the target file
    cat > $TARGET_BIN_DIR/$TARGET_FILENAME
    chmod +x $TARGET_BIN_DIR/$TARGET_FILENAME
    
    # Update PATH in .bashrc if not already present
    if ! grep -q \"$TARGET_BIN_DIR\" ~/.bashrc; then
        echo 'export PATH=\"\$PATH:$TARGET_BIN_DIR\"' >> ~/.bashrc
        echo 'PATH updated in .bashrc'
    fi
"

echo "Done. You can now run '$TARGET_FILENAME' inside the VM."