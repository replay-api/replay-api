# Stage 1: Build Environment
FROM ubuntu:latest AS builder

# Environment Variables (Set these when building the image)
ENV STEAM_USER ""
ENV STEAM_PASSWORD ""

# Install dependencies
RUN apt-get update && apt-get install -y \
    lib32gcc1 \
    libsdl2-2.0-0 \
    steamcmd \
    && rm -rf /var/lib/apt/lists/*

# Create a non-root user and group
RUN useradd -m -u 1000 steam
RUN groupadd -g 1000 steam
RUN usermod -aG steam steam

# Set up directories
WORKDIR /home/steam

# Download 
USER steam
RUN steamcmd \
    +login $STEAM_USER $STEAM_PASSWORD \
    +force_install_dir /home/steam/server \
    +app_update 730 validate \
    +quit

# Stage 2: Runtime Environment
FROM ubuntu:latest

# Set up directories
RUN mkdir -p /home/steam/server

# Copy only the necessary files from the builder stage
COPY --from=builder /home/steam/server /home/steam/server

# Create a non-root user and group (same as in builder stage)
RUN useradd -m -u 1000 steam
RUN groupadd -g 1000 steam
RUN usermod -aG steam steam

# Environment Variables for the final image
ENV CSGO_SERVER_NAME "My CS2 Server"
ENV CSGO_RCON_PASSWORD "your_strong_password"
ENV CSGO_GAME_MODE "competitive"
ENV CSGO_MAPGROUP "mg_active"
ENV CSGO_MAP "de_dust2"

# Copy configuration files
WORKDIR /home/steam/server
COPY server.cfg /home/steam/server/csgo/cfg/server.cfg
COPY start_server.sh /home/steam/server/

# Make the script executable
RUN chmod +x /home/steam/server/start_server.sh

# Switch to the non-root user
USER steam

# Expose game server ports (UDP and TCP)
EXPOSE 27015
EXPOSE 27015/udp

# Start the server on container launch
CMD ["/home/steam/server/start_server.sh"]
