#!/bin/bash

# VM Go Test Suite
# This script tests the vm-go implementation by creating VMs, verifying IPs, and testing SSH connectivity

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[0;33m'
NC='\033[0m'

# Test configuration
VM_GO_BINARY="./vm-go"
TEST_VM_PREFIX="test-go-vm"
TEST_COUNT=2
REPORT_FILE="test-report-$(date +%Y%m%d-%H%M%S).txt"

# Test results
TESTS_PASSED=0
TESTS_FAILED=0
TESTS_TOTAL=0

# Helper functions
log_test() {
    echo -e "${BLUE}[TEST]${NC} $1"
    echo "[TEST] $1" >> "$REPORT_FILE"
}

log_pass() {
    echo -e "${GREEN}[PASS]${NC} $1"
    echo "[PASS] $1" >> "$REPORT_FILE"
    ((TESTS_PASSED++))
    ((TESTS_TOTAL++))
}

log_fail() {
    echo -e "${RED}[FAIL]${NC} $1"
    echo "[FAIL] $1" >> "$REPORT_FILE"
    ((TESTS_FAILED++))
    ((TESTS_TOTAL++))
}

log_info() {
    echo -e "${YELLOW}[INFO]${NC} $1"
    echo "[INFO] $1" >> "$REPORT_FILE"
}

# Initialize report
echo "VM Go Test Suite Report" > "$REPORT_FILE"
echo "Date: $(date)" >> "$REPORT_FILE"
echo "======================================" >> "$REPORT_FILE"
echo "" >> "$REPORT_FILE"

# Test 1: Check if binary exists
log_test "Checking if vm-go binary exists"
if [ -f "$VM_GO_BINARY" ]; then
    log_pass "Binary found at $VM_GO_BINARY"
else
    log_fail "Binary not found at $VM_GO_BINARY"
    log_info "Building binary..."
    go build -o vm-go
    if [ -f "$VM_GO_BINARY" ]; then
        log_pass "Binary built successfully"
    else
        log_fail "Failed to build binary"
        exit 1
    fi
fi

# Test 2: Check if binary is executable
log_test "Checking if binary is executable"
if [ -x "$VM_GO_BINARY" ]; then
    log_pass "Binary is executable"
else
    log_fail "Binary is not executable"
    chmod +x "$VM_GO_BINARY"
    log_info "Made binary executable"
fi

# Test 3: Test help command
log_test "Testing help command"
if $VM_GO_BINARY --help &> /dev/null; then
    log_pass "Help command works"
else
    log_fail "Help command failed"
fi

# Test 4: Test list command (should work even with no VMs)
log_test "Testing list command"
if $VM_GO_BINARY list &> /dev/null; then
    log_pass "List command works"
else
    log_fail "List command failed"
fi

# Test 5: Create test VMs
log_info "Creating $TEST_COUNT test VMs..."
VM_NAMES=()
VM_IPS=()
VM_PORTS=()
VM_USERS=()
VM_PASSWORDS=()

for i in $(seq 1 $TEST_COUNT); do
    VM_NAME="${TEST_VM_PREFIX}-${i}"
    VM_NAMES+=("$VM_NAME")
    
    log_test "Creating VM: $VM_NAME"
    
    # Create VM with port forwarding (faster than bridge for testing)
    if $VM_GO_BINARY create --name "$VM_NAME" --net-type portfwd --memory 1G --disk 5G 2>&1 | tee -a "$REPORT_FILE"; then
        log_pass "VM $VM_NAME created"
        
        # Extract credentials from output (this is a simplified approach)
        # In a real scenario, we'd parse the info.json file
        VM_USERS+=("user01")
        VM_PASSWORDS+=("password")  # We'd need to extract this from the output or info.json
        
    else
        log_fail "Failed to create VM $VM_NAME"
        continue
    fi
    
    # Small delay between VM creations
    sleep 2
done

# Test 6: Verify VMs appear in list
log_test "Verifying VMs appear in list output"
LIST_OUTPUT=$($VM_GO_BINARY list)
for VM_NAME in "${VM_NAMES[@]}"; do
    if echo "$LIST_OUTPUT" | grep -q "$VM_NAME"; then
        log_pass "VM $VM_NAME appears in list"
    else
        log_fail "VM $VM_NAME does not appear in list"
    fi
done

# Test 7: Check VM status (should be running)
log_test "Checking if VMs are running"
for VM_NAME in "${VM_NAMES[@]}"; do
    if echo "$LIST_OUTPUT" | grep "$VM_NAME" | grep -q "RUNNING"; then
        log_pass "VM $VM_NAME is running"
    else
        log_fail "VM $VM_NAME is not running"
    fi
done

# Test 8: Extract SSH connection info
log_test "Extracting SSH connection information"
for VM_NAME in "${VM_NAMES[@]}"; do
    INFO_FILE="$HOME/.vm/vms/$VM_NAME/info.json"
    if [ -f "$INFO_FILE" ]; then
        log_pass "Found info.json for $VM_NAME"
        
        # Extract SSH port and credentials using jq if available, otherwise grep
        if command -v jq &> /dev/null; then
            SSH_PORT=$(jq -r '.ssh.port' "$INFO_FILE")
            USERNAME=$(jq -r '.username' "$INFO_FILE")
            PASSWORD=$(jq -r '.password' "$INFO_FILE")
        else
            SSH_PORT=$(grep -o '"port": [0-9]*' "$INFO_FILE" | cut -d' ' -f2)
            USERNAME=$(grep -o '"username": "[^"]*"' "$INFO_FILE" | cut -d'"' -f4)
            PASSWORD=$(grep -o '"password": "[^"]*"' "$INFO_FILE" | cut -d'"' -f4)
        fi
        
        log_info "VM $VM_NAME - SSH: 127.0.0.1:$SSH_PORT, User: $USERNAME"
        VM_PORTS+=("$SSH_PORT")
    else
        log_fail "info.json not found for $VM_NAME"
    fi
done

# Test 9: Wait for VMs to boot (cloud-init to complete)
log_info "Waiting for VMs to complete boot (60 seconds)..."
sleep 60

# Test 10: Test SSH connectivity
log_test "Testing SSH connectivity to VMs"
for i in "${!VM_NAMES[@]}"; do
    VM_NAME="${VM_NAMES[$i]}"
    SSH_PORT="${VM_PORTS[$i]}"
    USERNAME="${VM_USERS[$i]}"
    
    if [ -z "$SSH_PORT" ]; then
        log_fail "No SSH port found for $VM_NAME, skipping SSH test"
        continue
    fi
    
    log_info "Testing SSH to $VM_NAME on port $SSH_PORT..."
    
    # Test SSH connection (with StrictHostKeyChecking=no for testing)
    if timeout 10 ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null \
        -o ConnectTimeout=5 -p "$SSH_PORT" "$USERNAME@127.0.0.1" "echo 'SSH test successful'" 2>&1 | grep -q "SSH test successful"; then
        log_pass "SSH connection to $VM_NAME successful"
    else
        log_fail "SSH connection to $VM_NAME failed (VM may still be booting)"
    fi
done

# Test 11: Test stop command
log_test "Testing stop command"
if [ ${#VM_NAMES[@]} -gt 0 ]; then
    TEST_VM="${VM_NAMES[0]}"
    if $VM_GO_BINARY stop "$TEST_VM" 2>&1 | tee -a "$REPORT_FILE"; then
        log_pass "Stop command executed for $TEST_VM"
        
        # Verify VM is stopped
        sleep 2
        LIST_OUTPUT=$($VM_GO_BINARY list)
        if echo "$LIST_OUTPUT" | grep "$TEST_VM" | grep -q "STOPPED"; then
            log_pass "VM $TEST_VM is stopped"
        else
            log_fail "VM $TEST_VM is still running after stop command"
        fi
    else
        log_fail "Stop command failed for $TEST_VM"
    fi
fi

# Test 12: Cleanup - delete test VMs
log_info "Cleaning up test VMs..."
for VM_NAME in "${VM_NAMES[@]}"; do
    log_test "Deleting VM: $VM_NAME"
    if $VM_GO_BINARY delete "$VM_NAME" --force 2>&1 | tee -a "$REPORT_FILE"; then
        log_pass "VM $VM_NAME deleted"
    else
        log_fail "Failed to delete VM $VM_NAME"
    fi
done

# Test 13: Verify VMs are deleted
log_test "Verifying VMs are deleted"
LIST_OUTPUT=$($VM_GO_BINARY list)
for VM_NAME in "${VM_NAMES[@]}"; do
    if echo "$LIST_OUTPUT" | grep -q "$VM_NAME"; then
        log_fail "VM $VM_NAME still appears in list after deletion"
    else
        log_pass "VM $VM_NAME successfully removed"
    fi
done

# Generate final report
echo "" >> "$REPORT_FILE"
echo "======================================" >> "$REPORT_FILE"
echo "Test Summary" >> "$REPORT_FILE"
echo "======================================" >> "$REPORT_FILE"
echo "Total Tests: $TESTS_TOTAL" >> "$REPORT_FILE"
echo "Passed: $TESTS_PASSED" >> "$REPORT_FILE"
echo "Failed: $TESTS_FAILED" >> "$REPORT_FILE"
echo "" >> "$REPORT_FILE"

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "${GREEN}All tests passed!${NC}"
    echo "Status: ALL TESTS PASSED" >> "$REPORT_FILE"
    SUCCESS_RATE=100
else
    echo -e "${YELLOW}Some tests failed${NC}"
    echo "Status: SOME TESTS FAILED" >> "$REPORT_FILE"
    SUCCESS_RATE=$((TESTS_PASSED * 100 / TESTS_TOTAL))
fi

echo "Success Rate: ${SUCCESS_RATE}%" >> "$REPORT_FILE"
echo "" >> "$REPORT_FILE"
echo "Full report saved to: $REPORT_FILE"

# Display summary
echo ""
echo "======================================="
echo "Test Summary"
echo "======================================="
echo "Total Tests: $TESTS_TOTAL"
echo -e "Passed: ${GREEN}$TESTS_PASSED${NC}"
echo -e "Failed: ${RED}$TESTS_FAILED${NC}"
echo "Success Rate: ${SUCCESS_RATE}%"
echo ""
echo "Full report: $REPORT_FILE"

exit $TESTS_FAILED
