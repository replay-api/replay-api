# Security Policy

## Supported Versions

We are currently providing security updates for the following versions:

| Version | Supported          |
| ------- | ------------------ |
| 2.x     | :white_check_mark: |
| 1.x     | :white_check_mark: |
| < 1.0   | :x:                |

## Reporting a Vulnerability

Security is a priority for LeetGaming Replay API. If you discover a security vulnerability, we appreciate your help in disclosing it responsibly.

### How to Report

**DO NOT** create a public issue for security vulnerabilities. Instead:

1. **Send an email** to: `security@leetgaming.pro`
2. **Include**:
   - Detailed description of the vulnerability
   - Steps to reproduce
   - Potential impact
   - Suggestions for fixes (if any)

### What to Expect

- **Initial response**: Within 48 hours
- **Assessment**: Within 7 days
- **Update**: Every 7 days until resolution
- **Disclosure**: After the fix is released

### Responsible Disclosure Process

1. **Report**: Send the security report
2. **Confirmation**: You'll receive confirmation of receipt
3. **Investigation**: Our team investigates the vulnerability
4. **Fix**: We develop and test a fix
5. **Release**: We release the fix in a new version
6. **Disclosure**: We publish a security advisory (if necessary)

### Rewards

We currently do not offer a bug bounty program. However, we publicly acknowledge responsible contributors (with your permission) in our README and releases.

## Security Best Practices

### For Developers

- **Never** commit credentials or secrets
- Use environment variables for sensitive configurations
- Validate and sanitize all user inputs
- Use proper authentication and authorization
- Keep dependencies updated
- Follow security by design principles

### Security Checklist

Before committing:

- [ ] No hardcoded credentials
- [ ] User inputs validated
- [ ] Authorization verified for sensitive operations
- [ ] Dependencies updated
- [ ] Security tests executed
- [ ] Logs don't contain sensitive information

## Known Vulnerabilities

No known vulnerabilities at this time. If you discover one, please report it following the process above.

## Security History

### 2025-01-XX
- **CVE-XXXX-XXXX**: Description of fixed vulnerability
  - **Severity**: Critical/High/Medium/Low
  - **Fixed version**: 2.0.1
  - **Details**: [Link to security advisory]

## Contact

For security concerns:
- **Email**: `security@leetgaming.pro`
- **PGP Key**: [Add PGP key if available]

---

**Thank you for helping keep LeetGaming Replay API secure! 🔒**
