# See here for image contents: https://github.com/microsoft/vscode-dev-containers/tree/v0.158.0/containers/go/.devcontainer/base.Dockerfile

FROM golang:1.18

# Install go tools
RUN go install github.com/magefile/mage@latest

# Install Nodemon
ARG NODE_VERSION="12"
RUN curl -fsSL https://deb.nodesource.com/setup_$NODE_VERSION.x | bash - \
  && apt-get install -y nodejs \
  && npm i -g nodemon

WORKDIR /workspace