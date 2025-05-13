---
sidebar_position: 3
title: OIDC Authentication
---

:::info
Introduced in version 0.9.0
:::

# Setting Up OIDC Authentication

Arcane can integrate with an OpenID Connect (OIDC) provider to let users log in with their existing accounts from services like Google, Okta, Keycloak, or other compatible identity providers. This guide explains how to configure it.

## What is OIDC?

OIDC is a standard way for applications (like Arcane) to verify a user's identity using an external Identity Provider (IdP) without needing to handle passwords directly. It's a secure and common method for single sign-on (SSO).

## Before You Start

You'll need to have an OIDC provider set up and have registered Arcane as a client application with that provider. Your OIDC provider will give you several key pieces of information that Arcane needs:

- **Client ID**: A unique identifier for Arcane within your OIDC provider.
- **Client Secret**: A secret password that Arcane uses to authenticate itself to the OIDC provider. Keep this secure!
- **Authorization Endpoint URL**: The URL where users will be redirected to log in.
- **Token Endpoint URL**: The URL Arcane uses to exchange an authorization code for an access token.
- **Userinfo Endpoint URL**: The URL Arcane uses to get information about the logged-in user.
- **Scopes**: Defines what information Arcane can request about the user (e.g., `openid email profile`).

You will also need to configure a **Redirect URI** (also known as a Callback URL) in your OIDC provider. This is the URL where the provider will send the user back to Arcane after they log in. For Arcane, this is typically:
`http://your-arcane-address/auth/oidc/callback`
(e.g., `http://localhost:3000/auth/oidc/callback` or `https://arcane.example.com/auth/oidc/callback`)

## Configuring Arcane for OIDC

You'll primarily configure OIDC for Arcane using environment variables when running it in Docker.

### Environment Variables

These are the "sticky notes" you give to Arcane when it starts up to tell it about your OIDC provider.

- **`PUBLIC_OIDC_ENABLED`**: Set this to `true` to enable OIDC login.
  - Example: `PUBLIC_OIDC_ENABLED=true`
- **`OIDC_CLIENT_ID`**: The Client ID from your OIDC provider.
  - Example: `OIDC_CLIENT_ID="your_arcane_client_id_from_provider"`
- **`OIDC_CLIENT_SECRET`**: The Client Secret from your OIDC provider.
  - Example: `OIDC_CLIENT_SECRET="your_super_secret_client_secret_from_provider"`
- **`OIDC_REDIRECT_URI`**: The exact Redirect URI you configured in your OIDC provider for Arcane.
  - Example: `OIDC_REDIRECT_URI="http://localhost:3000/auth/oidc/callback"`
- **`OIDC_AUTHORIZATION_ENDPOINT`**: The Authorization Endpoint URL from your OIDC provider.
  - Example: `OIDC_AUTHORIZATION_ENDPOINT="https://your-provider.com/oauth2/authorize"`
- **`OIDC_TOKEN_ENDPOINT`**: The Token Endpoint URL from your OIDC provider.
  - Example: `OIDC_TOKEN_ENDPOINT="https://your-provider.com/oauth2/token"`
- **`OIDC_USERINFO_ENDPOINT`**: The Userinfo Endpoint URL from your OIDC provider.
  - Example: `OIDC_USERINFO_ENDPOINT="https://your-provider.com/oauth2/userinfo"`
- **`OIDC_SCOPES`**: The scopes Arcane should request. This is a space-separated list.
  - Default: `"openid email profile"`
  - Example: `OIDC_SCOPES="openid email profile groups"`

### Example `docker-compose.yml` Snippet

Here's how you might add these to your `docker-compose.yml` file:

```yaml
# In your docker-compose.yml file
services:
  arcane:
    # ... image, ports, volumes ...
    environment:
      # --- Standard Arcane Variables ---
      - APP_ENV=production
      - PUID=1000
      - PGID=1000
      - PUBLIC_SESSION_SECRET=your_super_long_random_secret_here

      # --- OIDC Configuration ---
      - PUBLIC_OIDC_ENABLED=true
      - OIDC_CLIENT_ID="your_arcane_client_id_from_provider"
      - OIDC_CLIENT_SECRET="your_super_secret_client_secret_from_provider"
      - OIDC_REDIRECT_URI="http://your-arcane-address/auth/oidc/callback" # IMPORTANT: Match your provider!
      - OIDC_AUTHORIZATION_ENDPOINT="https://your-provider.com/oauth2/authorize"
      - OIDC_TOKEN_ENDPOINT="https://your-provider.com/oauth2/token"
      - OIDC_USERINFO_ENDPOINT="https://your-provider.com/oauth2/userinfo"
      - OIDC_SCOPES="openid email profile" # Adjust as needed
    # ... other settings ...
```

## Enabling in Arcane UI

Once the environment variables are set and Arcane is restarted:

1.  Go to the Arcane web UI.
2.  Navigate to **Settings** -> **Authentication**.
3.  You should see options related to OIDC. Ensure "Enable OIDC" (or a similar option) is checked if it's not automatically enabled by `PUBLIC_OIDC_ENABLED=true`.
    _(Note: The exact UI option might vary or be controlled entirely by the environment variable)._

After configuration, you should see an option to "Login with OIDC" (or similar) on the Arcane login page.

## Troubleshooting

- **Double-check Redirect URI:** The most common issue is a mismatch between the `OIDC_REDIRECT_URI` in Arcane's environment variables and the Redirect URI configured in your OIDC provider. They must match exactly.
- **Provider Logs:** Check the logs from your OIDC provider for any error messages if login fails.
- **Arcane Logs:** Check Arcane's container logs for any OIDC-related errors.
- **Client Secret:** Ensure the `OIDC_CLIENT_SECRET` is correct and hasn't been accidentally changed or had special characters misinterpreted by your shell or Docker Compose file (quoting can help).
