#!/bin/bash

# Check if a version number is provided
if [ -z "$1" ]; then
    echo "Usage: $0 <version>"
    exit 1
fi

VERSION=$1

# Directories for each application
apps=("auth" "productlist" "userinfo" "webserver")

# Iterate over each application and build them
for app in "${apps[@]}"; do
    echo "Building $app for Docker scratch..."

    # Change to the application's directory
    cd "$app" || exit

    # Build the Go application statically
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o "$app"

    if [ $? -ne 0 ]; then
        echo "Build failed for $app"
        exit 1
    fi

    echo "$app built successfully."

    # Create a Dockerfile for the scratch image
    cat <<EOF > Dockerfile
# Use scratch image
FROM scratch

# Add application binary to the container
COPY $app /$app

# Run the application
ENTRYPOINT ["/$app"]
EOF

    # Build the Docker image
    docker build -t "${app}:${VERSION}" .

    if [ $? -ne 0 ]; then
        echo "Docker build failed for $app"
        exit 1
    fi

    echo "Docker image for $app built successfully: ${app}:${VERSION}"

    # Change back to the root directory
    cd - > /dev/null
done

# Version control - tagging the build
echo "Tagging the build with version $VERSION"
git add .
git commit -m "Build version $VERSION"
git tag "v$VERSION"
git push origin --tags

echo "Build, Docker image creation, and versioning completed successfully."
