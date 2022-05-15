FROM golang:1.18

# Install go tools
RUN go install github.com/magefile/mage@latest

# Install Nodemon
ARG NODE_VERSION="12"
RUN curl -fsSL https://deb.nodesource.com/setup_$NODE_VERSION.x | bash - \
  && apt-get update \
  && apt-get install -y nodejs libczmq-dev libzmq3-dev \
  && npm i -g nodemon

WORKDIR /workspace