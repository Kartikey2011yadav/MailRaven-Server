# Quickstart: Testing Standard Client Compliance

## Prerequisites
- MailRaven running (`go run main.go serve`)
- Email Client (Thunderbird, Outlook, or `openssl s_client`)

## Autodiscover Verification

**Mozilla Style**:
```bash
curl http://localhost:8080/.well-known/autoconfig/mail/config-v1.1.xml?emailaddress=test@example.com
```
*Expect: XML with IMAP/SMTP settings.*

**Microsoft Style**:
```bash
curl -X POST -d @request.xml http://localhost:8080/autodiscover/autodiscover.xml
```
*request.xml content:*
```xml
<?xml version="1.0" encoding="utf-8"?>
<Autodiscover xmlns="http://schemas.microsoft.com/exchange/autodiscover/outlook/requestschema/2006">
  <Request>
    <EMailAddress>test@example.com</EMailAddress>
    <AcceptableResponseSchema>http://schemas.microsoft.com/exchange/autodiscover/outlook/responseschema/2006a</AcceptableResponseSchema>
  </Request>
</Autodiscover>
```
*Expect: XML with IMAP settings.*

## IMAP Verification

1. **Connect**:
   ```bash
   openssl s_client -connect localhost:993 -crlf
   ```

2. **Login**:
   ```text
   A01 LOGIN user@example.com password
   ```

3. **Select Inbox**:
   ```text
   A02 SELECT INBOX
   ```
   *Expect: `* EXISTS`, `* RECENT`, `A02 OK`*

4. **Fetch Message**:
   ```text
   A03 FETCH 1 BODY[]
   ```
   *Expect: Raw email content.*

5. **IDLE (Push)**:
   ```text
   A04 IDLE
   ```
   *Expect: `+ idling`. Send an email to the user in another terminal. Expect `* EXISTS`.*
