# Work from kindes/base image for linux/amd64 and linux/arm64.
KINDBASE_IMAGE="kindest/base:v20210729-302b42d2"

# Our image name:tag for a containerd-in-Docker testing image.
CIND_IMAGE="lxkns/cind:inception"

# Name of test container instance.
CNTR_NAME="${CNTR_NAME:-containerd-in-docker}"

# While waiting for the containerd docker container to boot and start its
# inner testing container, print a "." every n lines of container log output.
LINESPERDOT=25
