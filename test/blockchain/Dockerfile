FROM node:20-alpine

# Install curl for healthchecks
RUN apk add --no-cache curl

WORKDIR /app

# Install dependencies
COPY package.json package-lock.json ./
RUN npm ci

# Copy contract files
COPY . .

# Compile contracts
RUN npx hardhat compile

EXPOSE 8545

CMD ["npx", "hardhat", "node", "--hostname", "0.0.0.0"]
