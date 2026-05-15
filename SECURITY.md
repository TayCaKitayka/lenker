# Security Policy

Lenker is an early-stage project and is not production-hardened yet.

Please do not publish exploit details, working proof-of-concept attacks, leaked
tokens, or private operational data in a public issue before maintainers have a
reasonable chance to review the report.

## Reporting Security Issues

Preferred reporting path:

1. Use GitHub private vulnerability reporting if it is enabled for the
   repository.
2. If private reporting is not available, contact the maintainers privately
   through the repository owner or project channels.
3. If no private channel is available, open a minimal public issue asking for a
   secure contact path. Do not include exploit details in that public issue.

Security reports should include:

- affected component;
- impact summary;
- reproduction steps, if safe to share privately;
- whether secrets, tokens, logs, or user data may be exposed;
- suggested fix or mitigation, if known.

## Security-Sensitive Areas

The following areas need especially careful review:

- admin authentication and sessions;
- node registration and bootstrap;
- mTLS, certificates, and trust material;
- future config signing and verification;
- secrets handling and log redaction;
- subscription keys and token storage;
- API authorization boundaries.

Marketplace and billing are outside the current MVP scope and should not be
reported as implemented security surfaces yet.

## Disclosure Expectations

The project cannot promise a formal SLA at this stage. The goal is coordinated,
practical handling of security issues without putting users or providers at
unnecessary risk.
