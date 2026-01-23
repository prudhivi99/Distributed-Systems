#!/bin/bash

#############################################
# Distributed Systems - Prerequisites Setup
# Run: chmod +x prereq.sh && ./prereq.sh
#############################################

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  Distributed Systems - Prerequisites  ${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""

#############################################
# Update System
#############################################
echo -e "${GREEN}[1/7] Updating system...${NC}"
sudo apt update && sudo apt upgrade -y

#############################################
# Install Basic Tools
#############################################
echo -e "${GREEN}[2/7] Installing basic tools...${NC}"
sudo apt install -y \
    curl wget git unzip jq make \
    build-essential ca-certificates gnupg \
    lsb-release apt-transport-https \
    software-properties-common \
    htop tree vim tmux \
    postgresql-client

#############################################
# Install Go 1.22
#############################################
echo -e "${GREEN}[3/7] Installing Go 1.22...${NC}"
GO_VERSION="1.22.5"
if ! command -v go &> /dev/null; then
    cd /tmp
    wget -q https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz
    sudo rm -rf /usr/local/go
    sudo tar -C /usr/local -xzf go${GO_VERSION}.linux-amd64.tar.gz
    rm go${GO_VERSION}.linux-amd64.tar.gz
fi

if ! grep -q '/usr/local/go/bin' ~/.bashrc; then
    echo '' >> ~/.bashrc
    echo '# Go environment' >> ~/.bashrc
    echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
    echo 'export GOPATH=$HOME/go' >> ~/.bashrc
    echo 'export PATH=$PATH:$GOPATH/bin' >> ~/.bashrc
fi

export PATH=$PATH:/usr/local/go/bin
export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin

echo -e "${GREEN}   Go: $(go version)${NC}"

#############################################
# Install Docker
#############################################
echo -e "${GREEN}[4/7] Installing Docker...${NC}"
if ! command -v docker &> /dev/null; then
    sudo apt remove -y docker docker-engine docker.io containerd runc 2>/dev/null || true
    sudo install -m 0755 -d /etc/apt/keyrings
    curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg
    sudo chmod a+r /etc/apt/keyrings/docker.gpg
    echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu $(. /etc/os-release && echo "$VERSION_CODENAME") stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
    sudo apt update
    sudo apt install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
fi

sudo systemctl start docker
sudo systemctl enable docker
sudo usermod -aG docker $USER

echo -e "${GREEN}   Docker: $(sudo docker --version)${NC}"

#############################################
# Install Docker Compose
#############################################
echo -e "${GREEN}[5/7] Installing Docker Compose...${NC}"
if ! command -v docker-compose &> /dev/null; then
    sudo curl -L "https://github.com/docker/compose/releases/download/v2.24.0/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
    sudo chmod +x /usr/local/bin/docker-compose
fi

echo -e "${GREEN}   Docker Compose: $(docker-compose --version)${NC}"

#############################################
# Install kubectl
#############################################
echo -e "${GREEN}[6/7] Installing kubectl...${NC}"
if ! command -v kubectl &> /dev/null; then
    curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
    sudo install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl
    rm kubectl
fi

#############################################
# Install Go Tools
#############################################
echo -e "${GREEN}[7/7] Installing Go tools...${NC}"
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest 2>/dev/null || true

#############################################
# Summary
#############################################
echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  Installation Complete!               ${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo -e "  ${GREEN}✓${NC} Go 1.22"
echo -e "  ${GREEN}✓${NC} Docker"
echo -e "  ${GREEN}✓${NC} Docker Compose"
echo -e "  ${GREEN}✓${NC} kubectl"
echo -e "  ${GREEN}✓${NC} PostgreSQL client"
echo -e "  ${GREEN}✓${NC} Basic tools (git, make, jq, tmux, vim)"
echo ""
echo -e "${YELLOW}>>> Run 'newgrp docker' or logout/login for Docker permissions${NC}"
echo ""
