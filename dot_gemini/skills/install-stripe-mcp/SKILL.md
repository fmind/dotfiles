---
name: install-stripe-mcp
description: Install the official Stripe MCP server in the current project's .gemini/settings.json so Gemini can manage customers, payments, products, prices, subscriptions, and invoices without leaving the session.
---

# Install Stripe MCP

Drops the official `@stripe/mcp` server into `.gemini/settings.json` for the current project. Use this when Stripe API operations (customers, charges, products, prices, subscriptions, invoices, refunds) happen regularly in the project.

## When to Trigger

- The repo imports `stripe` (Node) or `stripe` (Python), uses Stripe webhooks, or holds Stripe-bound config.
- The user wants Stripe API tools available in the main session.
- Verify first: `grep -q '"stripe"' .gemini/settings.json 2>/dev/null` — skip if already present.

## Install

Merge into `.gemini/settings.json` at the project root (create the file if missing):

```json
{
  "mcpServers": {
    "stripe": {
      "command": "npx",
      "args": [
        "-y",
        "@stripe/mcp",
        "--tools=all"
      ],
      "env": {
        "STRIPE_SECRET_KEY": "$STRIPE_SECRET_KEY"
      },
      "includeTools": []
    }
  }
}
```

Scope tools to a subset by replacing `--tools=all` with a comma list (e.g. `--tools=customers,payment_intents,invoices`). For Connect platforms, add `--stripe-account=acct_...`.

## Tool Filtering for Context Efficiency

MCP tool descriptions are loaded eagerly at session start. Pin `includeTools` to the handful of tools you actually use; excluded tools cost zero context tokens.

```text
/mcp desc stripe
```

## Authentication

Export a Stripe secret key (use a **restricted key** for CI/agent use; full secret keys grant total account access):

```bash
export STRIPE_SECRET_KEY=sk_test_...   # or rk_live_...
```

For experimentation, prefer a **test-mode** key (`sk_test_*`).

## Documentation

- [Stripe MCP docs](https://docs.stripe.com/mcp)
- [`@stripe/mcp` on npm](https://www.npmjs.com/package/@stripe/mcp)
- [Restricted API keys](https://docs.stripe.com/keys#create-restricted-api-secret-key)
- [Gemini CLI MCP servers reference](https://geminicli.com/docs/tools/mcp-server/)
