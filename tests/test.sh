set -e

# Run tests

for test in $(find tests -name '*.yml'); do
    echo "Running test: $test"
    echo "------------------------------------------------------"

    go run main.go -- perform -m test -c $test

    echo "------------------------------------------------------\n\n"
done
