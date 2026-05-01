---
name: install-google-pay-wallet-mcp
description: Install Google Pay & Wallet MCP into .gemini/settings.json. Use when Pay/Wallet integration is central to the project.
---

# Install Google Pay & Wallet MCP

Drops the Google Pay & Wallet Developer MCP server into `.gemini/settings.json` for the current project. Use this when Pay/Wallet integration work happens in nearly every session of the project — otherwise prefer the `pay-wallet` subagent (`~/.gemini/agents/pay-wallet.md`), which loads the MCP only when invoked and keeps the parent context clean.

## When to Trigger

- The project integrates Google Pay flows, mints Wallet pass JWTs, or manages issuer classes/objects.
- The user wants Pay/Wallet tools available in the main session without invoking the subagent.
- Verify first: `grep -q '"pay-wallet"' .gemini/settings.json 2>/dev/null` — skip if already present.

## Install

Merge into `.gemini/settings.json` at the project root (create the file if missing):

```json
{
  "mcpServers": {
    "pay-wallet": {
      "httpUrl": "https://paydeveloper.googleapis.com/mcp",
      "authProviderType": "google_credentials",
      "includeTools": []
    }
  }
}
```

## Tool Filtering for Context Efficiency

MCP tool descriptions are loaded eagerly at session start. Pin `includeTools` to the handful of tools you actually use; excluded tools cost zero context tokens.

```text
/mcp desc pay-wallet
```

## Authentication

Uses your default Google credentials. Run `gcloud auth application-default login` if not already authenticated.

## Companion Agent

The `pay-wallet` subagent (`~/.gemini/agents/pay-wallet.md`) wraps the same MCP and loads it lazily. Keep the subagent for cross-project ad-hoc work even after installing at the project level.

## Documentation

- [Pay & Wallet — MCP reference](https://developers.google.com/wallet/reference/mcp)
- [Google Pay API](https://developers.google.com/pay/api)
- [Google Wallet API](https://developers.google.com/wallet)
- [Google Cloud MCP supported products](https://docs.cloud.google.com/mcp/supported-products)
- [Gemini CLI MCP servers reference](https://geminicli.com/docs/tools/mcp-server/)
