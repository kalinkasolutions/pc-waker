# pc-waker

A tiny self-hosted web UI to wake PCs on your LAN via Wake-on-LAN (WoL).

- Single static Go binary in a `scratch` image (~10–15 MB).
- Hosts configured in an easy-to-edit YAML file mounted into the container.
- No database, no auth, no JavaScript framework.

## Quick start

```sh
cp config.example.yaml config.yaml
$EDITOR config.yaml          # add your machines' names + MAC addresses
docker compose up --build -d
```

Then open <http://localhost:8080> and click **Wake** next to a machine.

## Configuration

`config.yaml` is mounted read-only at `/config/config.yaml`:

```yaml
port: 8080                   # web UI port (env PORT overrides)
broadcast: 255.255.255.255   # default broadcast target
wol_port: 9                  # WoL UDP port (7 is also common)

hosts:
  - name: Desktop
    mac: AA:BB:CC:DD:EE:FF
  - name: Media PC
    mac: "11:22:33:44:55:66"
    broadcast: 192.168.1.255   # optional per-host override
```

Editing the config requires restarting the container (`docker compose restart`)
to reload it. Hosts with an unparseable MAC are logged and skipped at startup.

## Why `network_mode: host`?

Wake-on-LAN sends a UDP broadcast "magic packet". For that broadcast to reach
physical devices on your LAN, the container must share the host's network stack.
Docker's default bridge network does **not** forward broadcasts, so the compose
file uses `network_mode: host`. With host networking the `ports:` key is ignored
— the app binds `PORT` directly on the host.

## Prerequisites on the target PCs

- Enable **Wake-on-LAN** in the BIOS/UEFI and/or the OS network adapter settings.
- WoL generally works from sleep/hibernate and full shutdown (S5) depending on
  hardware; wired Ethernet is far more reliable than Wi-Fi.

## Security

There is no authentication. Only run this on a trusted LAN and do **not** expose
it to the internet. Put it behind a reverse proxy with auth if you need remote
access.

## Verifying without a sleeping PC

Confirm the packet is emitted by sniffing while you click Wake:

```sh
sudo tcpdump -i any -n udp port 9
```

You should see a 102-byte UDP broadcast to the configured address.
