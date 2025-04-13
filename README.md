Hey there,
I wrote this Go reverse proxy because I couldn't find anything that would have a similar behavior.
I wanted to run a java server on my home server but I soon found out that, even in idle, it doubled my power draw.
So this is my solution.
The reverse proxy will automatically turn on the server when someone tries to connect and will also shut it down 2 minutes after everyone disconnected (small race conditions may apply and might turn off a bit sooner, but only if the server is empty).

It is mostly intended to be used in docker compose
This is a sample of my config
```yaml
crafty:
  container_name: crafty
  image: registry.gitlab.com/crafty-controller/crafty-4:latest
  restart: always
  ports:
      - "8000:8000"
      - "8800:8800"
      - "8443:8443"
  volumes:
      - ./crafty/backups:/crafty/backups
      - ./crafty/logs:/crafty/logs
      - ./crafty/servers:/crafty/servers
      - ./crafty/config:/crafty/app/config
      - ./crafty/import:/crafty/import
craftyreverseproxy:
  image: andreicerbulescu/craftyreverseproxy:latest
  container_name: craftyreverseproxy
  ports:
    - "3120-3130:3120-3130"
  volumes:
    - ./craftyreverseproxy:/craftyproxy
  depends_on:
    crafty:
      condition: service_started
  restart: unless-stopped
```

You should create a folder in your folder called "craftyreverseproxy" and create a "config.yaml" file inside of it.
This is a sample of the config.yaml:
```yaml
api_url: "http://crafty:8443"
username: "admin"
password: "password"
timeout: "2m"
auto_shutdown: true
log_level: "INFO"

addresses:
  - crafty_host:
      addr: "crafty"
      port: 25565
    listener:
      addr: "localhost"
      port: 25565
    protocol: "tcp"
  - crafty_host:
      addr: "crafty"
      port: 25566
    listener:
      addr: "localhost"
      port: 25566
    protocol: "tcp"
```

For docker compose, the ip is the container's name. Make sure you expose the ports of the craftyreverseproxy in your docker-compose.

You will connect to the crafty server using your_ip:external_port. For instance 192.168.0.30:3120
I wanted to also implement UDP communication but it is a bit more trickier to auto shutdown the server.
