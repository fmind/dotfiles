# Tool Contracts

Field list for every tool exposed to a model, referenced from the [prompt-design](SKILL.md) workflow. Enforcement stays in runtime code; the description only tells the model how to choose and call the tool.

- **Name and purpose**: unique, action-oriented name and one concise purpose sentence.
- **Prerequisites**: state or authority that must hold before the call.
- **Arguments**: required and optional arguments with types, units, defaults, and validation rules.
- **Side effects and idempotency**: what the call changes and whether a retry is safe.
- **Authorization and consent**: which calls need user confirmation and what evidence of consent looks like.
- **Timeout and retry**: budget per call and the retry policy the runtime applies.
- **Errors**: each error's meaning and the model's expected reaction (retry, ask, stop).
- **Sensitive return fields**: values that must not be echoed, logged, or forwarded.
- **Success evidence**: the observable result that proves the call did what it claims.
