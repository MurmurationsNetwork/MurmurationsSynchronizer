#!/bin/bash

# Exit immediately if a command exits with a non-zero status.
set -e

# Setup environment variables.
ssh_private_key="$SSH_PRIVATE_KEY"
server_ip="$SERVER_IP"
kubeconfig_path="$KUBECONFIG_PATH"

# Transform the string into valid JSON and then parse it.
formatted_json=$(echo $EXCLUDE_MATRIX | \
    sed 's/service: \([^,}]*\)/"service": "\1"/g')
exclude_services=($(echo $formatted_json | jq -r '.[] | .service'))
echo "Excluded services: ${exclude_services[*]}"

# Setup SSH.
echo "Setting up SSH..."
mkdir -p ~/.ssh
echo "$ssh_private_key" > ssh_key
chmod 600 ssh_key
eval $(ssh-agent -s)
ssh-add ssh_key
ssh-keyscan -H "$server_ip" >> ~/.ssh/known_hosts

# Copy Kubernetes config from the server.
echo "Copying Kubernetes configuration..."
scp "root@$server_ip:$kubeconfig_path" ./kubeconfig

# Setting KUBECONFIG environment variable.
export KUBECONFIG=./kubeconfig

# Replace localhost IP in Kubeconfig.
sed -i 's/https:\/\/127.0.0.1:6443/https:\/\/'$server_ip':6443/' \
    ./kubeconfig

# Install kubectl.
echo "Installing kubectl..."
curl -LO "https://storage.googleapis.com/kubernetes-release/release/\
$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)\
/bin/linux/amd64/kubectl"
chmod +x ./kubectl
sudo mv ./kubectl /usr/local/bin/kubectl

# Deployment logic for each service.
make deploy

# Clean up.
echo "Cleaning up..."
eval $(ssh-agent -k)
rm ssh_key