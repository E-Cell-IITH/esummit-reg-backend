#!/usr/bin/env bash

# The local Go main file, sql and environment file
GO_MAIN_FILE="/Users/bhaskarmandal/Desktop/aa45/esummit-reg-backend/cmd/api/main.go"
ENV_FILE=".env"

if [ -f ${ENV_FILE} ]; then
    set -o allexport
    source ${ENV_FILE}
    set +o allexport
fi

VM_PATH="/home/${VM_USER}/esummit-server" 
SERVICE_NAME="esummit-server"

if [ -z "${VM_USER}" ] || [ -z "${VM_HOST}" ] || [ -z "${VM_PATH}" ] || [ -z "${SERVICE_NAME}" ]; then
  echo "Please configure the variables in the script before running it."
  exit 1
fi


echo "Building Go binary for Linux amd64..."
GOARCH=amd64 GOOS=linux go mod tidy
GOARCH=amd64 GOOS=linux go build -o "${SERVICE_NAME}" "${GO_MAIN_FILE}"

if [ $? -ne 0 ]; then
  echo "Go build failed. Exiting..."
  exit 1
fi

echo "Creating directory on the VM..."
ssh "${VM_USER}@${VM_HOST}" "mkdir -p ${VM_PATH}" "mkdir -p ${VM_PATH}/templates"

echo "Copying files to the VM..."
scp "${SERVICE_NAME}" "${VM_USER}@${VM_HOST}:${VM_PATH}/${SERVICE_NAME}"
scp "${ENV_FILE}" "${VM_USER}@${VM_HOST}:${VM_PATH}/.env"

scp -r templates \
    "${VM_USER}@${VM_HOST}:${VM_PATH}"

echo "Moving binary to /usr/local/bin..."
ssh "${VM_USER}@${VM_HOST}" << EOF
  sudo mv "${VM_PATH}/${SERVICE_NAME}" /usr/local/bin/${SERVICE_NAME}
  sudo chmod +x /usr/local/bin/${SERVICE_NAME}
EOF

SERVICE_FILE_CONTENT="[Unit]
Description=Go Server
After=network.target

[Service]
Type=simple
User=${VM_USER}
Group=${VM_USER}
WorkingDirectory=${VM_PATH}
EnvironmentFile=${VM_PATH}/.env
ExecStart=/usr/local/bin/${SERVICE_NAME}
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
"

echo "Creating systemd service file locally..."
echo "${SERVICE_FILE_CONTENT}" > "${SERVICE_NAME}.service"

echo "Copying systemd service file to the VM..."
scp "${SERVICE_NAME}.service" "${VM_USER}@${VM_HOST}:/tmp/${SERVICE_NAME}.service"

echo "Setting up systemd service on the VM..."
ssh "${VM_USER}@${VM_HOST}" << EOF
  sudo mv /tmp/${SERVICE_NAME}.service /etc/systemd/system/${SERVICE_NAME}.service
  sudo systemctl daemon-reload

  if [ -f ${VM_PATH}/.env ]; then
    set -o allexport
    source ${VM_PATH}/.env
    set +o allexport
  fi

  sudo systemctl enable ${SERVICE_NAME}
  sudo systemctl restart ${SERVICE_NAME}
EOF

echo "Cleaning up local artifacts..."
rm -f "${SERVICE_NAME}.service"

echo "Deployment complete!"

