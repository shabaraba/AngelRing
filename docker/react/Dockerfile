FROM node:latest
COPY ./react-app /app/react
WORKDIR /app
RUN npm install -g npm@latest
RUN npm install -g bun
RUN bunx create-vite react
