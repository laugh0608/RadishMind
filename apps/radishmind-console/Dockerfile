FROM node:24-alpine AS build

WORKDIR /workspace

COPY apps/radishmind-console/package.json apps/radishmind-console/package-lock.json apps/radishmind-console/
RUN cd apps/radishmind-console && npm ci

COPY apps/radishmind-console/ apps/radishmind-console/
COPY contracts/typescript/ contracts/typescript/

ARG VITE_RADISHMIND_PLATFORM_BASE_URL=http://127.0.0.1:7000
ENV VITE_RADISHMIND_PLATFORM_BASE_URL=${VITE_RADISHMIND_PLATFORM_BASE_URL}

RUN cd apps/radishmind-console && npm run build

FROM nginx:1.27-alpine AS runtime

COPY apps/radishmind-console/nginx.local.conf /etc/nginx/conf.d/default.conf
COPY --from=build /workspace/apps/radishmind-console/dist/ /usr/share/nginx/html/

EXPOSE 80
