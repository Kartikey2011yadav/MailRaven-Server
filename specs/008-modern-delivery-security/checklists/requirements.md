# Requirements Checklist

- [ ] **MTA-STS Policy**: `GET /.well-known/mta-sts.txt` returns valid policy.
- [ ] **MTA-STS Subdomain**: Requests to `mta-sts.example.com` are routed correctly.
- [ ] **TLS-RPT Endpoint**: `POST /.well-known/tlsrpt` accepts JSON.
- [ ] **TLS-RPT Storage**: Reports are saved to `tls_reports` table.
- [ ] **DANE Logic**: Outbound SMTP checks TLSA records if DNSSEC is valid.
- [ ] **DANE Fallback**: System behavior on DANE failure respects "Advisory" vs "Mandatory" levels (currently mandatory if TLSA exists).
