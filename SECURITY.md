# Security Policy

## Supported Development Target

This policy applies to the AviCosystem Backend repository.

## Reporting a Vulnerability

Do not publish sensitive vulnerability details, credentials, tokens, or exploit information in public issues.

Contact the repository maintainer privately through the contact information published on the repository profile.

Include:

- Affected component
- Impact
- Reproduction steps
- Suggested mitigation, if available

## Credential Handling

Never commit:

- `.env`
- Database passwords
- JWT secrets
- Redis credentials
- Private keys
- Production access tokens

## Response

Reports should be reviewed and prioritized according to their security impact.
