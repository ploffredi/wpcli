#!/bin/bash

# Set the path to the wpcli executable
WPCLI="./wpcli"

# Function to run a test and check the result
run_test() {
    local description=$1
    local command=$2
    local expected_exit_code=${3:-0}  # Default to 0 if not specified

    echo "=== TEST: $description ==="
    echo "Command: $command"
    echo "Expected exit code: $expected_exit_code"
    echo "Output:"

    # Run the command and capture the output
    output=$($command 2>&1)
    exit_code=$?

    # Display the output
    echo "$output"

    # Check if the command succeeded or failed as expected
    if [ $exit_code -eq $expected_exit_code ]; then
        echo "✅ Test PASSED (exit code: $exit_code)"
    else
        echo "❌ Test FAILED (exit code: $exit_code, expected: $expected_exit_code)"
    fi

    echo "================================"
    echo ""
}

# Build the wpcli executable
echo "Building wpcli executable..."
cd "$(dirname "$0")/.." || exit 1
go build
if [ $? -ne 0 ]; then
    echo "Failed to build wpcli executable"
    exit 1
fi
cd - > /dev/null || exit 1

# Test pkg install command - Success cases
run_test "Install the latest version of a package" "$WPCLI pkg install my-package"
run_test "Install a specific version of a package" "$WPCLI pkg install my-package --version 1.2.3"
run_test "Force install a package" "$WPCLI pkg install my-package --force"
run_test "Install a specific version with force flag" "$WPCLI pkg install my-package --version 1.2.3 --force"

# Test pkg install command - Error cases
run_test "Install without package name (missing required argument)" "$WPCLI pkg install" 1
run_test "Install with invalid flag" "$WPCLI pkg install my-package --invalid-flag" 1

# Test pkg remove command - Success cases
run_test "Remove a package" "$WPCLI pkg remove my-package"
run_test "Remove a package and its configuration files" "$WPCLI pkg remove my-package --purge"

# Test pkg remove command - Error cases
run_test "Remove without package name (missing required argument)" "$WPCLI pkg remove" 1
run_test "Remove with invalid flag" "$WPCLI pkg remove my-package --invalid-flag" 1

# Test pkg list command - Success cases
run_test "List only installed packages" "$WPCLI pkg list"
run_test "List all packages including uninstalled ones" "$WPCLI pkg list --all"
run_test "List packages in JSON format" "$WPCLI pkg list --format json"
run_test "List packages in YAML format" "$WPCLI pkg list --format yaml"
run_test "List all packages in table format" "$WPCLI pkg list --all --format table"

# Test pkg list command - Error cases
run_test "List with invalid format" "$WPCLI pkg list --format invalid" 1
run_test "List with invalid flag" "$WPCLI pkg list --invalid-flag" 1

# Test greet command - Success cases
run_test "Print a default greeting" "$WPCLI greet"
run_test "Print a greeting with a name" "$WPCLI greet John"
run_test "Print a greeting in Italian" "$WPCLI greet --language it"
run_test "Print a formal greeting in Spanish" "$WPCLI greet --language es --formal"
run_test "Print a formal greeting with a name in Italian" "$WPCLI greet Maria --language it --formal"

# Test greet command - Error cases
run_test "Greet with invalid language" "$WPCLI greet --language invalid" 1
run_test "Greet with invalid flag" "$WPCLI greet --invalid-flag" 1

# Test help commands
run_test "Show help for pkg install command" "$WPCLI pkg install --help"
run_test "Show help for greet command" "$WPCLI greet --help"
run_test "Show help for pkg command" "$WPCLI pkg --help"
run_test "Show general help" "$WPCLI --help"

# Test invalid commands
run_test "Invalid command" "$WPCLI invalid-command" 1
run_test "Invalid subcommand" "$WPCLI pkg invalid-subcommand" 0

echo "All tests completed!"
