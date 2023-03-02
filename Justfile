EXAMPLES := `ls examples/*`

build: fmt
    go build

fmt:
    go fmt

examples: build
    #!/usr/bin/env bash
    for d in {{EXAMPLES}}; do
        echo "Applying example $$d"
        terraform -chdir="$d" init
        terraform -chdir="$d" apply -auto-approve
    done
