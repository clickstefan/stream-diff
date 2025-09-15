#!/bin/bash

# Integration tests for Stream-Diff
# Tests the complete CLI functionality with real test data

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
BUILD_DIR="$PROJECT_ROOT/build"
BINARY="$BUILD_DIR/stream-diff"
TEST_OUTPUT_DIR="$PROJECT_ROOT/test-output"

# Test counters
TESTS_PASSED=0
TESTS_FAILED=0
TOTAL_TESTS=0

# Cleanup function
cleanup() {
    echo -e "${BLUE}Cleaning up test artifacts...${NC}"
    rm -rf "$TEST_OUTPUT_DIR"
}

# Set up cleanup trap
trap cleanup EXIT

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[PASS]${NC} $1"
}

log_error() {
    echo -e "${RED}[FAIL]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

# Test execution function
run_test() {
    local test_name="$1"
    local test_command="$2"
    local expected_exit_code="${3:-0}"
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    log_info "Running test: $test_name"
    
    if eval "$test_command"; then
        local exit_code=$?
        if [ $exit_code -eq $expected_exit_code ]; then
            log_success "$test_name"
            TESTS_PASSED=$((TESTS_PASSED + 1))
            return 0
        else
            log_error "$test_name (expected exit code $expected_exit_code, got $exit_code)"
            TESTS_FAILED=$((TESTS_FAILED + 1))
            return 1
        fi
    else
        local exit_code=$?
        if [ $exit_code -eq $expected_exit_code ]; then
            log_success "$test_name"
            TESTS_PASSED=$((TESTS_PASSED + 1))
            return 0
        else
            log_error "$test_name (expected exit code $expected_exit_code, got $exit_code)"
            TESTS_FAILED=$((TESTS_FAILED + 1))
            return 1
        fi
    fi
}

# Initialize test environment
init_tests() {
    log_info "Initializing integration tests..."
    
    # Create test output directory
    mkdir -p "$TEST_OUTPUT_DIR"
    
    # Check if binary exists
    if [ ! -f "$BINARY" ]; then
        log_error "Binary not found at $BINARY. Please build first with 'make build'"
        exit 1
    fi
    
    log_success "Test environment initialized"
}

# Test basic CLI functionality
test_cli_basic() {
    log_info "=== Testing Basic CLI Functionality ==="
    
    # Test help command
    run_test "CLI Help" "$BINARY --help >/dev/null"
    
    # Test version command
    run_test "Version Command" "$BINARY version >/dev/null"
    
    # Test invalid command
    run_test "Invalid Command" "$BINARY invalid-command >/dev/null 2>&1" 1
}

# Test validation functionality
test_validation() {
    log_info "=== Testing Configuration Validation ==="
    
    # Test valid configuration
    run_test "Validate Valid Config" \
        "$BINARY validate $PROJECT_ROOT/testdata/testcase1_simple_csv/config1.yaml >/dev/null"
    
    # Test multiple valid configurations
    run_test "Validate Multiple Configs" \
        "$BINARY validate $PROJECT_ROOT/testdata/testcase1_simple_csv/config1.yaml $PROJECT_ROOT/testdata/testcase1_simple_csv/config2.yaml >/dev/null"
    
    # Test non-existent file
    run_test "Validate Non-existent File" \
        "$BINARY validate non-existent.yaml >/dev/null 2>&1" 1
    
    # Test validation with explain flag
    run_test "Validate with Explain" \
        "$BINARY validate --explain $PROJECT_ROOT/testdata/testcase1_simple_csv/config1.yaml >/dev/null"
}

# Test comparison functionality
test_comparison() {
    log_info "=== Testing Data Comparison ==="
    
    # Test schema-only comparison
    run_test "Schema-only Comparison" \
        "$BINARY compare --schema-only $PROJECT_ROOT/testdata/testcase1_simple_csv/config1.yaml $PROJECT_ROOT/testdata/testcase1_simple_csv/config2.yaml > $TEST_OUTPUT_DIR/schema_comparison.yaml"
    
    # Test comparison with output file
    run_test "Comparison with Output File" \
        "$BINARY compare --schema-only --output $TEST_OUTPUT_DIR/comparison_result.yaml $PROJECT_ROOT/testdata/testcase1_simple_csv/config1.yaml $PROJECT_ROOT/testdata/testcase1_simple_csv/config2.yaml"
    
    # Verify output file was created
    run_test "Verify Output File Created" \
        "[ -f $TEST_OUTPUT_DIR/comparison_result.yaml ]"
    
    # Test comparison with sample size limit
    run_test "Comparison with Sample Size" \
        "$BINARY compare --schema-only --sample-size 100 $PROJECT_ROOT/testdata/testcase1_simple_csv/config1.yaml $PROJECT_ROOT/testdata/testcase1_simple_csv/config2.yaml >/dev/null"
    
    # Test comparison with verbose logging
    run_test "Comparison with Verbose Logging" \
        "$BINARY compare --verbose --schema-only $PROJECT_ROOT/testdata/testcase1_simple_csv/config1.yaml $PROJECT_ROOT/testdata/testcase1_simple_csv/config2.yaml >/dev/null 2>&1"
}

# Test different data source types
test_data_sources() {
    log_info "=== Testing Different Data Source Types ==="
    
    # Test CSV comparison
    run_test "CSV Data Source" \
        "$BINARY compare --schema-only $PROJECT_ROOT/testdata/testcase1_simple_csv/config1.yaml $PROJECT_ROOT/testdata/testcase1_simple_csv/config2.yaml >/dev/null"
    
    # Test JSON data source if available
    if [ -d "$PROJECT_ROOT/testdata/testcase2_nested_json" ]; then
        run_test "JSON Data Source" \
            "$BINARY compare --schema-only $PROJECT_ROOT/testdata/testcase2_nested_json/config1.yaml $PROJECT_ROOT/testdata/testcase2_nested_json/config2.yaml >/dev/null"
    fi
    
    # Test CSV with JSON strings if available
    if [ -d "$PROJECT_ROOT/testdata/testcase3_csv_with_json" ]; then
        run_test "CSV with JSON Strings" \
            "$BINARY compare --schema-only $PROJECT_ROOT/testdata/testcase3_csv_with_json/config1.yaml $PROJECT_ROOT/testdata/testcase3_csv_with_json/config2.yaml >/dev/null"
    fi
}

# Test error handling
test_error_handling() {
    log_info "=== Testing Error Handling ==="
    
    # Test missing configuration files
    run_test "Missing Config File" \
        "$BINARY compare missing1.yaml missing2.yaml >/dev/null 2>&1" 1
    
    # Test invalid YAML
    echo "invalid: yaml: content:" > "$TEST_OUTPUT_DIR/invalid.yaml"
    run_test "Invalid YAML Syntax" \
        "$BINARY validate $TEST_OUTPUT_DIR/invalid.yaml >/dev/null 2>&1" 1
    
    # Test insufficient arguments
    run_test "Compare Insufficient Args" \
        "$BINARY compare $PROJECT_ROOT/testdata/testcase1_simple_csv/config1.yaml >/dev/null 2>&1" 1
}

# Test AI insights and recommendations
test_ai_features() {
    log_info "=== Testing AI Features ==="
    
    # Test AI insights in comparison
    run_test "AI Insights in Comparison" \
        "$BINARY compare --schema-only $PROJECT_ROOT/testdata/testcase1_simple_csv/config1.yaml $PROJECT_ROOT/testdata/testcase1_simple_csv/config2.yaml | grep -q 'ai_insights'"
    
    # Test AI recommendations in validation
    run_test "AI Recommendations in Validation" \
        "$BINARY validate $PROJECT_ROOT/testdata/testcase1_simple_csv/config1.yaml | grep -q 'AI Recommendations'"
}

# Test output formats
test_output_formats() {
    log_info "=== Testing Output Formats ==="
    
    # Test YAML output (default)
    run_test "YAML Output Format" \
        "$BINARY compare --schema-only --format yaml $PROJECT_ROOT/testdata/testcase1_simple_csv/config1.yaml $PROJECT_ROOT/testdata/testcase1_simple_csv/config2.yaml >/dev/null"
    
    # Test unsupported format
    run_test "Unsupported Output Format" \
        "$BINARY compare --schema-only --format xml $PROJECT_ROOT/testdata/testcase1_simple_csv/config1.yaml $PROJECT_ROOT/testdata/testcase1_simple_csv/config2.yaml >/dev/null 2>&1" 1
}

# Performance tests
test_performance() {
    log_info "=== Testing Performance Features ==="
    
    # Test with sample size limitation
    run_test "Performance - Sample Size Limit" \
        "timeout 30 $BINARY compare --schema-only --sample-size 10 $PROJECT_ROOT/testdata/testcase1_simple_csv/config1.yaml $PROJECT_ROOT/testdata/testcase1_simple_csv/config2.yaml >/dev/null"
    
    # Test timeout handling (command should complete quickly with small sample)
    run_test "Performance - Quick Execution" \
        "timeout 10 $BINARY compare --schema-only --sample-size 1 $PROJECT_ROOT/testdata/testcase1_simple_csv/config1.yaml $PROJECT_ROOT/testdata/testcase1_simple_csv/config2.yaml >/dev/null"
}

# Main test execution
main() {
    log_info "Starting Stream-Diff Integration Tests"
    log_info "Binary: $BINARY"
    log_info "Project Root: $PROJECT_ROOT"
    
    init_tests
    
    # Run all test suites
    test_cli_basic
    test_validation
    test_comparison
    test_data_sources
    test_error_handling
    test_ai_features
    test_output_formats
    test_performance
    
    # Print test summary
    echo
    log_info "=== Test Summary ==="
    echo "Total Tests: $TOTAL_TESTS"
    echo -e "Passed: ${GREEN}$TESTS_PASSED${NC}"
    echo -e "Failed: ${RED}$TESTS_FAILED${NC}"
    
    if [ $TESTS_FAILED -eq 0 ]; then
        log_success "All integration tests passed! 🎉"
        exit 0
    else
        log_error "Some tests failed! ❌"
        exit 1
    fi
}

# Run main function
main "$@"