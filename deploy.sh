#!/bin/bash
# SwiftLog Deployment Script
# This script helps deploy SwiftLog to a remote server

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Default values
ENVIRONMENT="production"
VERSION="latest"
DEPLOY_PATH="/opt/swiftlog"

# Parse arguments
while [[ $# -gt 0 ]]; do
  case $1 in
    -h|--host)
      DEPLOY_HOST="$2"
      shift 2
      ;;
    -u|--user)
      DEPLOY_USER="$2"
      shift 2
      ;;
    -k|--key)
      SSH_KEY="$2"
      shift 2
      ;;
    -v|--version)
      VERSION="$2"
      shift 2
      ;;
    -p|--path)
      DEPLOY_PATH="$2"
      shift 2
      ;;
    -e|--env)
      ENVIRONMENT="$2"
      shift 2
      ;;
    --help)
      echo "SwiftLog Deployment Script"
      echo ""
      echo "Usage: $0 [OPTIONS]"
      echo ""
      echo "Options:"
      echo "  -h, --host HOST       Deployment host (required)"
      echo "  -u, --user USER       SSH user (required)"
      echo "  -k, --key KEY_PATH    SSH private key path (optional)"
      echo "  -v, --version VER     Docker image version (default: latest)"
      echo "  -p, --path PATH       Deployment path on server (default: /opt/swiftlog)"
      echo "  -e, --env ENV         Environment name (default: production)"
      echo "  --help                Show this help message"
      echo ""
      echo "Example:"
      echo "  $0 -h server.example.com -u deploy -v v0.1.0"
      exit 0
      ;;
    *)
      echo -e "${RED}Unknown option: $1${NC}"
      echo "Use --help for usage information"
      exit 1
      ;;
  esac
done

# Validate required parameters
if [ -z "$DEPLOY_HOST" ] || [ -z "$DEPLOY_USER" ]; then
  echo -e "${RED}Error: Host and user are required${NC}"
  echo "Use --help for usage information"
  exit 1
fi

# Build SSH command
SSH_CMD="ssh"
SCP_CMD="scp"
if [ -n "$SSH_KEY" ]; then
  SSH_CMD="ssh -i $SSH_KEY"
  SCP_CMD="scp -i $SSH_KEY"
fi

echo -e "${GREEN}🚀 SwiftLog Deployment${NC}"
echo "=================================="
echo "Environment: $ENVIRONMENT"
echo "Version:     $VERSION"
echo "Host:        $DEPLOY_HOST"
echo "User:        $DEPLOY_USER"
echo "Path:        $DEPLOY_PATH"
echo "=================================="
echo ""

# Create temporary deployment directory
TEMP_DIR=$(mktemp -d)
trap "rm -rf $TEMP_DIR" EXIT

echo -e "${YELLOW}📦 Preparing deployment package...${NC}"

# Copy necessary files
cp docker-compose.yaml $TEMP_DIR/
cp loki-config.yaml $TEMP_DIR/
cp .env.example $TEMP_DIR/.env.template

# Modify docker-compose to use published images
cat docker-compose.yaml | \
  sed 's|build:.*|# build removed for deployment|g' | \
  sed 's|context:.*|# context removed|g' | \
  sed 's|dockerfile:.*|# dockerfile removed|g' \
  > $TEMP_DIR/docker-compose.yaml

# Determine image registry and repository
if [ "$VERSION" != "latest" ]; then
  IMAGE_TAG="$VERSION"
else
  IMAGE_TAG="latest"
fi

# Add image references to docker-compose
sed -i.bak "/ingestor:/a\\
    image: ghcr.io/\${GITHUB_REPOSITORY:-aliancn/swiftlog}/ingestor:${IMAGE_TAG}" $TEMP_DIR/docker-compose.yaml

sed -i.bak "/api:/a\\
    image: ghcr.io/\${GITHUB_REPOSITORY:-aliancn/swiftlog}/api:${IMAGE_TAG}" $TEMP_DIR/docker-compose.yaml

sed -i.bak "/websocket:/a\\
    image: ghcr.io/\${GITHUB_REPOSITORY:-aliancn/swiftlog}/websocket:${IMAGE_TAG}" $TEMP_DIR/docker-compose.yaml

sed -i.bak "/ai-worker:/a\\
    image: ghcr.io/\${GITHUB_REPOSITORY:-aliancn/swiftlog}/ai-worker:${IMAGE_TAG}" $TEMP_DIR/docker-compose.yaml

sed -i.bak "/frontend:/a\\
    image: ghcr.io/\${GITHUB_REPOSITORY:-aliancn/swiftlog}/frontend:${IMAGE_TAG}" $TEMP_DIR/docker-compose.yaml

rm $TEMP_DIR/docker-compose.yaml.bak

# Create update script
cat > $TEMP_DIR/update.sh << 'EOFSCRIPT'
#!/bin/bash
set -e

echo "🚀 Updating SwiftLog services..."

# Load environment variables
if [ -f .env ]; then
  export $(cat .env | grep -v '^#' | xargs)
else
  echo "⚠️  Warning: .env file not found. Using defaults."
fi

# Pull latest images
echo "📥 Pulling Docker images..."
docker compose pull

# Stop and remove old containers
echo "🛑 Stopping old containers..."
docker compose down

# Start services
echo "🚀 Starting services..."
docker compose up -d --remove-orphans

# Wait for services
echo "⏳ Waiting for services to start..."
sleep 15

# Check status
echo "📊 Service status:"
docker compose ps

echo ""
echo "✅ Update complete!"
echo ""
echo "🌐 Access points:"
echo "   Frontend:  http://$(hostname -I | awk '{print $1}'):3000"
echo "   API:       http://$(hostname -I | awk '{print $1}'):8080"
echo "   WebSocket: ws://$(hostname -I | awk '{print $1}'):8081"
echo "   gRPC:      $(hostname -I | awk '{print $1}'):50051"
EOFSCRIPT

chmod +x $TEMP_DIR/update.sh

# Copy update-env script if exists
if [ -f "scripts/update-env.sh" ]; then
  cp scripts/update-env.sh $TEMP_DIR/
  chmod +x $TEMP_DIR/update-env.sh
fi

# Create package
echo -e "${YELLOW}📦 Creating deployment package...${NC}"
cd $TEMP_DIR
tar czf deploy.tar.gz *
cd - > /dev/null

# Copy to server
echo -e "${YELLOW}📤 Copying files to server...${NC}"
$SCP_CMD $TEMP_DIR/deploy.tar.gz $DEPLOY_USER@$DEPLOY_HOST:/tmp/

# Deploy on server
echo -e "${YELLOW}🚀 Deploying on server...${NC}"
$SSH_CMD $DEPLOY_USER@$DEPLOY_HOST << EOFSSH
set -e

# Extract package
cd /tmp
tar xzf deploy.tar.gz

# Create deployment directory
sudo mkdir -p $DEPLOY_PATH

# Backup existing .env
if [ -f $DEPLOY_PATH/.env ]; then
  sudo cp $DEPLOY_PATH/.env /tmp/.env.backup
  echo "✅ Backed up existing .env"
fi

# Copy files
sudo cp docker-compose.yaml $DEPLOY_PATH/
sudo cp loki-config.yaml $DEPLOY_PATH/
sudo cp update.sh $DEPLOY_PATH/
sudo chmod +x $DEPLOY_PATH/update.sh

# Copy update-env script if exists
if [ -f update-env.sh ]; then
  sudo cp update-env.sh $DEPLOY_PATH/
  sudo chmod +x $DEPLOY_PATH/update-env.sh
fi

# Restore or create .env
if [ -f /tmp/.env.backup ]; then
  sudo cp /tmp/.env.backup $DEPLOY_PATH/.env
  echo "✅ Restored .env from backup"
else
  sudo cp .env.template $DEPLOY_PATH/.env
  echo "⚠️  Created new .env from template."
fi

# Update environment variables from local environment if available
if [ -f $DEPLOY_PATH/update-env.sh ]; then
  echo ""
  echo "🔧 You can update environment variables by setting them in your shell:"
  echo "   export POSTGRES_PASSWORD=your-password"
  echo "   export JWT_SECRET=your-secret"
  echo "   export ADMIN_PASSWORD=your-admin-password"
  echo ""
  read -p "Do you want to update environment variables now? (y/N): " update_env
  if [ "$update_env" = "y" ] || [ "$update_env" = "Y" ]; then
    echo ""
    echo "Enter values (or press Enter to skip):"

    read -p "POSTGRES_PASSWORD: " input_postgres_password
    read -p "JWT_SECRET: " input_jwt_secret
    read -p "ADMIN_PASSWORD: " input_admin_password

    # Export for update-env.sh
    [ -n "$input_postgres_password" ] && export POSTGRES_PASSWORD="$input_postgres_password"
    [ -n "$input_jwt_secret" ] && export JWT_SECRET="$input_jwt_secret"
    [ -n "$input_admin_password" ] && export ADMIN_PASSWORD="$input_admin_password"

    cd $DEPLOY_PATH
    sudo -E ./update-env.sh .env
    cd - > /dev/null
  fi
fi

# Set ownership
sudo chown -R $DEPLOY_USER:$DEPLOY_USER $DEPLOY_PATH

# Run update
cd $DEPLOY_PATH
./update.sh

# Cleanup
rm -f /tmp/deploy.tar.gz /tmp/.env.backup
rm -f /tmp/*.yaml /tmp/update.sh /tmp/.env.template

echo ""
echo "✅ Deployment completed successfully!"
EOFSSH

echo ""
echo -e "${GREEN}✅ Deployment completed successfully!${NC}"
echo ""
echo "Next steps:"
echo "  1. Verify services are running: $SSH_CMD $DEPLOY_USER@$DEPLOY_HOST 'cd $DEPLOY_PATH && docker compose ps'"
echo "  2. View logs: $SSH_CMD $DEPLOY_USER@$DEPLOY_HOST 'cd $DEPLOY_PATH && docker compose logs -f'"
echo "  3. Configure CLI: ./cli/swiftlog config set --token YOUR_TOKEN --server $DEPLOY_HOST:50051"
