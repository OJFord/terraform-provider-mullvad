EXAMPLES := `ls examples/*`

build: fmt
    go build

debug: fmt
    #!/usr/bin/env bash
    set -eEuo pipefail
    function build() {
        echo "Building for debug"
        go build -gcflags='all=-N -l'
    }
    build
    trap build SIGUSR1
    gdlv -r stdout:/dev/stdout exec ./terraform-provider-mullvad -debug

fmt:
    go fmt

examples: build
    #!/usr/bin/env bash
    for d in {{EXAMPLES}}; do
        echo "Applying example $$d"
        terraform -chdir="$d" init
        terraform -chdir="$d" apply -auto-approve
    done
