FROM node:20-alpine

# Set working directory
WORKDIR /app

# Copy package files first (for better Docker layer caching)
COPY package.json package-lock.json ./

# Install dependencies (use ci for reproducible builds)
RUN npm ci --only=production

# Copy the rest of the application files
COPY . .

# Create a non-root user for security best practices
RUN addgroup -g 1001 -S nodejs && \
    adduser -S nodejs -u 1001 && \
    chown -R nodejs:nodejs /app

# Switch to non-root user
USER nodejs

# Set NODE_ENV to production
ENV NODE_ENV=production

# Command to run the application
# This will be executed by the Kubernetes CronJob
CMD ["node", "index.js"]