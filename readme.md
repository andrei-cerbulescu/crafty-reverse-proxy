Hey there,
I wrote this Go reverse proxy because I couldn't find anything that would have a similar behavior.
I wanted to run a java server on my home server but I soon found out that, even in idle, it doubled my power draw.
So this is my solution.
The reverse proxy will automatically turn on the server when someone tries to connect and will also shut it down 2 minutes after everyone disconnected (small race conditions may apply and might turn off a bit sooner, but only if the server is empty).

It is mostly intended to be used in docker compose
This is a sample of my config
```
craftyreverseproxy:
  image: andreicerbulescu/craftyreverseproxy:latest
  container_name: craftyreverseproxy
  ports:
    - "3120-3130:3120-3130"
  volumes:
    - ./craftyreverseproxy:/craftyproxy
  restart: unless-stopped
```

You should add a config in the /craftyreverseproxy folder "config.json".
This is a sample of the config.json:
```
{
  "api_url": "https://crafty_container:8443",
  "username": "user",
  "password": "pass",
  "addresses": [
    {
      "internal_ip": "crafty",
      "internal_port": "25565",
      "external_ip": "craftyreverseproxy",
      "external_port": "3120",
      "protocol": "tcp",
      "Others": []
    }
  ]
}
```
I wanted to also implement UDP communication but it is a bit more trickier to auto shutdown the server.
