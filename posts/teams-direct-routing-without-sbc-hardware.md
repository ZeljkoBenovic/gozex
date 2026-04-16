# Microsoft Teams Direct Routing Without the Hardware SBC

**tsbc** is a containerised Teams Direct Routing SBC using Kamailio and RTPEngine — SIP/UDP PBX to Microsoft Teams via Docker, without hardware appliances.

If your company runs Microsoft Teams and you want to use your own SIP trunk instead of paying for Microsoft Calling Plans, you need Direct Routing — and Direct Routing requires a Session Border Controller.

The problem? Hardware SBCs from vendors like AudioCodes, Ribbon, or Oracle are expensive, need dedicated rack space, and introduce yet another appliance to patch and manage. For small-to-medium deployments, the cost-to-complexity ratio is hard to justify.

Enter **tsbc** — a fully containerised Session Border Controller that bridges any SIP/UDP-based PBX or SIP trunk directly into Microsoft Teams Direct Routing. No hardware. No appliance licence. Just Docker.

## What tsbc Does

tsbc orchestrates three battle-tested open-source components into a single deployable stack:

- **Kamailio** — handles SIP signalling, registration, and routing
- **RTPEngine** — manages Real-Time Protocol (RTP) media transcoding and relay
- **Let's Encrypt / Certbot** — provisions and auto-renews the TLS certificate that Teams requires

Together they cover the full SBC function: signalling normalisation, codec negotiation, NAT traversal, and TLS termination — without a single line of proprietary firmware.

## Why Bother With a Software SBC?

Traditional SBC hardware solves real problems: media relay, SIP normalisation between carrier and Microsoft, TLS enforcement. Software solutions like tsbc solve exactly the same problems at a fraction of the cost:

- **No capital expenditure** — runs on any Linux host or container platform
- **Teams compliance** — meets the TLS/SRTP requirements Microsoft mandates
- **Full control** — modify Kamailio routing logic to match your specific trunk topology
- **Disaster recovery** — recreate the entire stack from a single `docker compose up`

## Quick Deployment

The simplest deployment needs just a public-facing Linux host with ports 5060/5061 (SIP) and the RTP media port range open.

Clone the repository and copy the example config:

```bash
git clone https://github.com/ZeljkoBenovic/tsbc
cd tsbc
cp .env.example .env
```

Edit `.env` with your domain, SIP trunk credentials, and Teams tenant details:

```env
DOMAIN=sbc.yourdomain.com
SIP_TRUNK_HOST=sip.yourprovider.com
SIP_TRUNK_PORT=5060
TEAMS_TENANT=yourtenant.onmicrosoft.com
LETSENCRYPT_EMAIL=admin@yourdomain.com
```

Bring the stack up:

```bash
docker compose up -d
```

Certbot handles the initial TLS certificate request automatically. Once DNS propagates and the certificate is issued, Kamailio starts accepting connections from the Teams infrastructure.

## Configuring Teams Direct Routing

On the Microsoft 365 side, register your SBC domain in the Teams admin centre and add a voice route pointing at it:

```powershell
# Register the SBC
New-CsOnlinePSTNGateway -Fqdn sbc.yourdomain.com -SipSignalingPort 5061 -Enabled $true

# Create a voice route
New-CsOnlineVoiceRoute -Name "PSTN-Route" -NumberPattern "^\+[1-9]\d{6,14}$" `
  -OnlinePstnGatewayList sbc.yourdomain.com
```

Assign the routing policy to a user and make a test call. The first ring confirms the full SIP-to-Teams path is working.

## Production Considerations

For production deployments, a few additional steps are worth taking:

**Media port range**: Configure RTPEngine to use a predictable UDP range (e.g. 20000–30000) and open that range in your firewall. Teams sends media directly to RTPEngine, bypassing Kamailio entirely once the call is established.

**High availability**: Place two tsbc instances behind a load balancer or DNS failover. Since RTPEngine carries media state, active-active HA requires shared state — for most SME deployments, active-passive is sufficient.

**Monitoring**: Kamailio exposes XMLRPC metrics; RTPEngine has a JSON control socket. Both integrate cleanly with Prometheus via community exporters.

## Conclusion

tsbc makes Teams Direct Routing accessible without the appliance tax. If you already manage Linux hosts and have Docker available, you have everything you need to run a compliant, production-grade SBC.

Deploy [tsbc from GitHub](https://github.com/ZeljkoBenovic/tsbc) and run Teams Direct Routing without the hardware.
