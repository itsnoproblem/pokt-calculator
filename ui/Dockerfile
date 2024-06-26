FROM node:14-alpine AS builder
WORKDIR /app
COPY package.json .
COPY yarn.lock .
RUN yarn install

COPY . .
RUN yarn build

COPY ./env.sh .
COPY .env.* .

# Add bash
RUN apk add --no-cache bash

# Make our shell script executable
RUN chmod +x env.sh

CMD [ "/bin/bash/", "-c", "./env.sh" ]

FROM nginx:1.21.0-alpine as runtime
COPY --from=builder /app/build /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80

# Copy .env file and shell script to container
WORKDIR /usr/share/nginx/html

# Add bash
RUN apk add --no-cache bash

CMD ["/bin/bash", "-c", "nginx -g \"daemon off;\""]

